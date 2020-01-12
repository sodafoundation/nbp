// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package block

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/opensds/nbp/csi/common"
	"github.com/opensds/nbp/csi/util"
	"github.com/opensds/opensds/client"
	nbputil "github.com/opensds/nbp/util"	
	"github.com/opensds/opensds/contrib/connector"
	"github.com/opensds/opensds/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Volume struct {
	Client *client.Client
}

func NewVolume(c *client.Client) *Volume {
	return &Volume{Client: c}
}

// CreateVolume implementation
func (v *Volume) CreateVolume(req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	// build volume body
	volumebody := &model.VolumeSpec{}
	volumebody.Name = req.GetName()
	var secondaryAZ = util.OpensdsDefaultSecondaryAZ
	var enableReplication = false
	var attachMode = "rw"
	glog.V(5).Infof("create volume parameters %+v", req.GetParameters())
	for k, v := range req.GetParameters() {
		switch k {
		case common.ParamProfile:
			if v == "" {
				msg := "profile id cannot be empty"
				glog.Error(msg)
				return nil, status.Error(codes.InvalidArgument, msg)
			}
			volumebody.ProfileId = v
		case common.ParamEnableReplication:
			if strings.ToLower(v) == "true" {
				enableReplication = true
			}
		case common.ParamSecondaryAZ:
			secondaryAZ = v
		case common.PublishAttachMode:
			if strings.ToLower(v) == "ro" {
				attachMode = "ro"
			}
		}
	}

	size := common.GetSize(req.GetCapacityRange())
	volumebody.Size = size

	if req.GetAccessibilityRequirements() != nil {
		volumebody.AvailabilityZone = common.GetZone(req.GetAccessibilityRequirements(), TopologyZoneKey)
	}

	glog.V(5).Infof("volume body: %+v", volumebody)

	volExist, err := v.FindVolume(volumebody.Name)
	if err != nil {
		return nil, err
	}

	if volExist == nil {
		volExist, err = v.Client.CreateVolume(volumebody)
		if err != nil {
			msg := fmt.Sprintf("create volume failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	}

	glog.V(5).Info("waiting until volume is created")
	volStable, err := common.WaitForStatusStable(volExist.Id, func(id string) (interface{}, error) {
		return v.Client.GetVolume(id)
	})

	if err != nil {
		msg := fmt.Sprintf("failed to create volume: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	vol := volStable.(*model.VolumeSpec)
	// return volume info
	volumeinfo := &csi.Volume{
		CapacityBytes: vol.Size * util.GiB,
		VolumeId:      vol.Id,
		VolumeContext: map[string]string{
			VolumeName:               vol.Name,
			VolumeStatus:             vol.Status,
			VolumeAZ:                 vol.AvailabilityZone,
			VolumePoolId:             vol.PoolId,
			VolumeProfileId:          vol.ProfileId,
			VolumeLvPath:             vol.Metadata["lvPath"],
			common.PublishAttachMode: attachMode,
		},

		AccessibleTopology: []*csi.Topology{
			{
				Segments: map[string]string{
					TopologyZoneKey: volumebody.AvailabilityZone,
				},
			},
		},
	}

	glog.V(5).Infof("response volume info = %+v", volumeinfo)

	if enableReplication && volExist == nil {
		volumebody.AvailabilityZone = secondaryAZ
		volumebody.Name = SecondaryPrefix + req.Name
		sVol, err := v.Client.CreateVolume(volumebody)
		if err != nil {
			msg := fmt.Sprintf("failed to create second volume: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		_, err = common.WaitForStatusStable(sVol.Id, func(id string) (interface{}, error) {
			return v.Client.GetVolume(id)
		})

		if err != nil {
			msg := fmt.Sprintf("failed to create volume: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		replicaBody := &model.ReplicationSpec{
			Name:              req.Name,
			PrimaryVolumeId:   vol.Id,
			SecondaryVolumeId: sVol.Id,
			ReplicationMode:   model.ReplicationModeSync,
			ReplicationPeriod: 0,
		}
		replicaResp, err := v.Client.CreateReplication(replicaBody)
		if err != nil {
			msg := fmt.Sprintf("create replication failed: %v", err)
			glog.Errorf(msg)
			return nil, status.Error(codes.Internal, msg)
		}
		volumeinfo.VolumeContext[VolumeReplicationId] = replicaResp.Id
	}

	return &csi.CreateVolumeResponse{
		Volume: volumeinfo,
	}, nil
}

// FindVolume implementation
func (v *Volume) FindVolume(volName string) (*model.VolumeSpec, error) {
	volumes, err := v.Client.ListVolumes()
	if err != nil {
		msg := fmt.Sprintf("list volumes failed: %v", err)
		glog.Error(msg)
		return nil, errors.New(msg)
	}

	for _, volume := range volumes {
		if volume.Name == volName {
			return volume, nil
		}
	}

	return nil, nil
}

// DeleteVolume implementation
func (v *Volume) DeleteVolume(volId string) (*csi.DeleteVolumeResponse, error) {
	vol, _ := v.Client.GetVolume(volId)
	if vol == nil {
		return &csi.DeleteVolumeResponse{}, nil
	}

	r := v.getReplicationByVolume(volId)
	if r != nil {
		if err := v.Client.DeleteReplication(r.Id, nil); err != nil {
			msg := fmt.Sprintf("delete replication failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
		if err := v.Client.DeleteVolume(r.PrimaryVolumeId, &model.VolumeSpec{}); err != nil {
			msg := fmt.Sprintf("delete primary volume failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
		if err := v.Client.DeleteVolume(r.SecondaryVolumeId, &model.VolumeSpec{}); err != nil {
			msg := fmt.Sprintf("delete secondary volume failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
	} else {
		if err := v.Client.DeleteVolume(volId, &model.VolumeSpec{}); err != nil {
			msg := fmt.Sprintf("delete volume failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ExpandVolume implementation
func (v *Volume) ExpandVolume(req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ControllerExpandVolume volume ID must be provided")
	}
	glog.V(5).Infof("expand volume called for volume id: %v", volumeID)

	extendVolBody := &model.ExtendVolumeSpec{}
	extendVolBody.NewSize = common.GetSize(req.GetCapacityRange())

	volSpec, err := v.Client.ExtendVolume(volumeID, extendVolBody)
	if err != nil {
		msg := fmt.Sprintf("extend volume failed: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	_, err = common.WaitForStatusStable(volSpec.Id, func(id string) (interface{}, error) {
		return v.Client.GetVolume(id)
	})

	if err != nil {
		msg := fmt.Sprintf("failed to extend volume: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	glog.V(5).Infof("expand volume succeded with new capacity(GB): %v", extendVolBody.NewSize)

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         extendVolBody.NewSize * util.GiB,
		NodeExpansionRequired: true,
	}, nil
}

// getReplicationByVolume implementation
func (v *Volume) getReplicationByVolume(volId string) *model.ReplicationSpec {
	replications, _ := v.Client.ListReplications()
	for _, r := range replications {
		if volId == r.PrimaryVolumeId || volId == r.SecondaryVolumeId {
			return r
		}
	}
	return nil
}

// ControllerPublishVolume implementation
func (v *Volume) ControllerPublishVolume(req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	attachMode, ok := req.VolumeContext[common.PublishAttachMode]
	if !ok {
		glog.Info("attach mode will use default value: rw")
		attachMode = "rw"
	}

	//check volume is exist
	volSpec, err := v.Client.GetVolume(req.VolumeId)
	if err != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s does not exist: %v",
			req.VolumeId, err)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	pool, err := v.Client.GetPool(volSpec.PoolId)
	if err != nil || pool == nil {
		msg := fmt.Sprintf("the pool %s does not exist: %v",
			volSpec.PoolId, err)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	var protocol = strings.ToLower(pool.Extras.IOConnectivity.AccessProtocol)
	attachReq := &model.VolumeAttachmentSpec{
		VolumeId:       req.VolumeId,
		HostId:         req.NodeId,
		AttachMode:     attachMode,
		ConnectionInfo: model.ConnectionInfo{
			DriverVolumeType: protocol,
		},
	}

	mode := req.VolumeCapability.AccessMode.Mode
	canAtMultiNode := false

	if csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER == mode {
		canAtMultiNode = true
	}

	err = v.isVolumeCanBePublished(canAtMultiNode, attachReq, volSpec.MultiAttach)
	if err != nil {
		return nil, err
	}

	newAttachment, errAttach := v.Client.CreateVolumeAttachment(attachReq)
	if errAttach != nil {
		msg := fmt.Sprintf("the volume %s is failed to be published to node %s, error info: %v",
			req.VolumeId, req.NodeId, errAttach)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	resp := &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			common.PublishHostIp:       connector.GetHostIP(),
			common.PublishHostId:       newAttachment.HostId,
			common.PublishAttachId:     newAttachment.Id,
			common.PublishAttachStatus: newAttachment.Status,
			common.PublishAttachMode:   attachMode,
		},
	}

	if replicationId, ok := req.VolumeContext[VolumeReplicationId]; ok {
		r, err := v.Client.GetReplication(replicationId)
		if err != nil {
			msg := fmt.Sprintf("failed to get replication: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		attachReq.VolumeId = r.SecondaryVolumeId

		secondaryVolume, err := v.Client.GetVolume(attachReq.VolumeId)
		if err != nil {
			msg := fmt.Sprintf("failed to get secondary volume: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		err = v.isVolumeCanBePublished(canAtMultiNode, attachReq, secondaryVolume.MultiAttach)
		if err != nil {
			return nil, err
		}

		newAttachment, errAttach := v.Client.CreateVolumeAttachment(attachReq)
		if errAttach != nil {
			msg := fmt.Sprintf("the volume %s failed to be published to node %s, error info %v",
				req.VolumeId, req.NodeId, errAttach)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
		resp.PublishContext[common.PublishSecondaryAttachId] = newAttachment.Id
	}

	return resp, nil
}

// isVolumePublished Check if the volume is published and compatible
func (v *Volume) isVolumeCanBePublished(canAtMultiNode bool, attachReq *model.VolumeAttachmentSpec,
	volMultiAttach bool) error {

	glog.V(5).Infof("start to check if volume can be published, canAtMultiNode = %v, attachReq = %v",
		canAtMultiNode, attachReq)

	attachments, err := v.Client.ListVolumeAttachments()
	if err != nil {
		msg := fmt.Sprintf("list volume attachments failed: %v", err)
		glog.Error(msg)
		return status.Error(codes.FailedPrecondition, msg)
	}

	msg := fmt.Sprintf("the volume %s can be published", attachReq.VolumeId)

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == attachReq.VolumeId {
			if attachSpec.HostId == attachReq.HostId {
				msg := fmt.Sprintf("the volume %s is publishing to the current node %s and no need to publish again", attachReq.VolumeId, attachReq.HostId)
				glog.Infof(msg)
				return nil
			}
			if !canAtMultiNode {
				msg := fmt.Sprintf("the volume %s has been published to the node %s and kubernetes does not have MULTI_NODE volume capability", attachReq.VolumeId, attachSpec.HostId)
				glog.Error(msg)
				return status.Error(codes.FailedPrecondition, msg)
			}
			if !volMultiAttach {
				msg := fmt.Sprintf("the volume %s has been published to the node %s, but the volume does not enable multiattach", attachReq.VolumeId, attachSpec.HostId)
				glog.Error(msg)
				return status.Error(codes.FailedPrecondition, msg)
			}
		}
	}

	glog.Info(msg)
	return nil
}

// ControllerUnpublishVolume implementation
func (v *Volume) ControllerUnpublishVolume(req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	//check volume is exist
	volSpec, errVol := v.Client.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s does not exist: %v", req.VolumeId, errVol)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := v.Client.ListVolumeAttachments()
	if err != nil {
		msg := fmt.Sprintf("failed to list volume attachments: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	var acts []*model.VolumeAttachmentSpec

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == req.VolumeId && attachSpec.HostId == req.NodeId {
			acts = append(acts, attachSpec)
			break
		}
	}

	if r := v.getReplicationByVolume(req.VolumeId); r != nil {
		for _, attachSpec := range attachments {
			if attachSpec.VolumeId == r.SecondaryVolumeId && attachSpec.HostId == req.NodeId {
				acts = append(acts, attachSpec)
				break
			}
		}
	}

	for _, act := range acts {
		if ok := common.UnpublishAttachmentList.IsExist(act.Id); !ok {
			glog.Infof("add attachment id %s into unpublish attachment list", act.Id)
			common.UnpublishAttachmentList.Add(act)
			common.UnpublishAttachmentList.PrintVolAttachList()
		}
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ListVolumes implementation
func (v *Volume) ListVolumes(req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	// only support list all the volumes at present
	volumes, err := v.Client.ListVolumes()
	if err != nil {
		return nil, err
	}

	ens := []*csi.ListVolumesResponse_Entry{}
	for _, v := range volumes {
		if v != nil {

			volumeinfo := &csi.Volume{
				CapacityBytes: v.Size,
				VolumeId:      v.Id,
				VolumeContext: map[string]string{
					"Name":             v.Name,
					"Status":           v.Status,
					"AvailabilityZone": v.AvailabilityZone,
					"PoolId":           v.PoolId,
					"ProfileId":        v.ProfileId,
				},
			}

			ens = append(ens, &csi.ListVolumesResponse_Entry{
				Volume: volumeinfo,
			})
		}
	}

	return &csi.ListVolumesResponse{
		Entries: ens,
	}, nil
}

// NodeStageVolume implementation
func (v *Volume) NodeStageVolume(req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	volId := req.VolumeId
	attachmentId := req.PublishContext[common.PublishAttachId]

	if r := v.getReplicationByVolume(volId); r != nil {
		if r.ReplicationStatus == model.ReplicationFailover {
			volId = r.SecondaryVolumeId
			attachmentId = req.PublishContext[common.PublishSecondaryAttachId]
		}
		if r.Metadata == nil {
			r.Metadata = make(map[string]string)
		}
		r.Metadata[AttachedVolumeId] = volId
		if _, err := v.Client.UpdateReplication(r.Id, r); err != nil {
			msg := fmt.Sprintf("update replication %s failed: %v", r.Id, err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	vol, attachment, err := v.getVolumeAndAttachment(volId, attachmentId)
	if nil != err {
		return nil, err
	}

	device := attachment.Mountpoint
	mountpoint := req.StagingTargetPath

	if 0 == len(device) || "-" == device {
		volConnector := connector.NewConnector(attachment.DriverVolumeType)
		if nil == volConnector {
			msg := fmt.Sprintf("unsupport driver volume type: %s", attachment.DriverVolumeType)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		devicePath, err := volConnector.Attach(attachment.ConnectionData)
		if nil != err || 0 == len(devicePath) || "-" == devicePath {
			msg := fmt.Sprintf("failed to find device: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		device = devicePath
	}

	mnt := req.VolumeCapability.GetMount()
	block := req.VolumeCapability.GetBlock()
	glog.V(5).Infof("volume capability Mount=%+v, Block=%+v\n", mnt, block)

	if nil != mnt && nil != block {
		msg := "volume mode cannot be both Block and Filesystem"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if nil == vol.Metadata {
		vol.Metadata = make(map[string]string)
	}

	if nil == block {
		vol.Metadata[CSIVolumeMode] = "Filesystem"
		// Format
		hopeFSType := "ext4"
		if mnt.FsType != "" {
			glog.Infof("use system fsType %s", mnt.FsType)
			hopeFSType = mnt.FsType
		}

		curFSType, err := connector.GetFSType(device)
		if err != nil {
			msg := err.Error()
			glog.Error(msg)
			return nil, status.Error(codes.Aborted, msg)
		}

		if curFSType == "" {
			if err := connector.Format(device, hopeFSType); err != nil {
				msg := fmt.Sprintf("failed to mkfs: %v", err)
				glog.Error(msg)
				return nil, status.Error(codes.Aborted, msg)
			}
		} else {
			glog.Infof("device: %s has been formatted yet, fsType: %s", device, curFSType)
		}

		// Mount
		mounted, err := connector.IsMounted(mountpoint)
		if err != nil {
			msg := fmt.Sprintf("failed to check mounted: %v", err)
			glog.Errorf(msg)
			return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
		}

		if mounted {
			glog.Info("volume is already mounted.")
			return &csi.NodeStageVolumeResponse{}, nil
		}

		glog.Info("mounting...")

		err = connector.Mount(device, mountpoint, hopeFSType, mnt.MountFlags)
		if err != nil {
			msg := fmt.Sprintf("failed to mount: %v", err)
			glog.Errorf(msg)
			return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
		}
	} else {
		vol.Metadata[CSIVolumeMode] = "Block"
		err = CreateSymlink(device, mountpoint)

		if err != nil {
			msg := fmt.Sprintf("failed to create a link: oldname=%v, newname=%v, err %v", device, mountpoint, err)
			glog.Error(msg)
			return nil, status.Error(codes.Aborted, msg)
		}
	}

	_, err = v.Client.UpdateVolume(vol.Id, vol)
	if err != nil {
		msg := fmt.Sprintf("update volume failed: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	glog.V(5).Info("node stage volume success")
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume implementation
func (v *Volume) NodeUnstageVolume(req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {

	vol, _, err := v.getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if CSIFilesystem == vol.Metadata[CSIVolumeMode] {
		// check volume is unmounted
		mounted, err := connector.IsMounted(req.StagingTargetPath)
		if err != nil {
			msg := fmt.Sprintf("check volume is unmounted failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		if !mounted {
			glog.Info("target path is already unmounted")
			return &csi.NodeUnstageVolumeResponse{}, nil
		}

		// Umount
		glog.Infof("node unpublish volume mountpoint: %s", req.StagingTargetPath)
		err = connector.Umount(req.StagingTargetPath)
		if err != nil {
			msg := fmt.Sprintf("failed to umount, %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	if CSIBlock == vol.Metadata[CSIVolumeMode] {
		_, err = exec.Command("rm", "-rf", req.StagingTargetPath).CombinedOutput()
		if err != nil {
			msg := fmt.Sprintf("rm -rf %v failed: %v", req.StagingTargetPath, err)
			glog.Errorf(msg)
			return nil, errors.New(msg)
		}
	}

	glog.V(5).Info("node unstage volume success")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume implementation
func (v *Volume) NodePublishVolume(req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	volId := req.VolumeId
	attachmentId := req.PublishContext[common.PublishAttachId]

	if r := v.getReplicationByVolume(volId); r != nil {
		volId = r.Metadata[AttachedVolumeId]
		attachmentId = r.Metadata[AttachedId]
	}

	_, _, err := v.getVolumeAndAttachment(volId, attachmentId)
	if nil != err {
		return nil, err
	}

	device := req.StagingTargetPath
	mountpoint := req.TargetPath

	mnt := req.VolumeCapability.GetMount()
	block := req.VolumeCapability.GetBlock()
	glog.V(5).Infof("volume capability mount=%+v, block=%+v\n", mnt, block)

	if nil != mnt && nil != block {
		msg := "volume mode cannot be both Block or Filesystem"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if nil == block {
		// Bind mount
		mountFlags := append(mnt.MountFlags, "bind")
		glog.V(5).Infof("req.Readonly, %v", req.Readonly)
		if req.Readonly {
			mountFlags = append(mountFlags, "ro")
		}

		fsType := "ext4"
		if mnt.FsType != "" {
			glog.Infof("use system fsType %s", mnt.FsType)
			fsType = mnt.FsType
		}

		// Mount
		mounted, err := connector.IsMounted(mountpoint)
		if err != nil {
			msg := fmt.Sprintf("failed to check mounted: %v", err)
			glog.Errorf(msg)
			return nil, status.Errorf(codes.FailedPrecondition, msg)
		}

		if mounted {
			glog.Info("volume is already mounted")
			return &csi.NodePublishVolumeResponse{}, nil
		}

		glog.Info("mounting...")

		err = connector.Mount(device, mountpoint, fsType, mountFlags)
		if err != nil {
			msg := fmt.Sprintf("failed to mount: %v", err)
			glog.Errorf(msg)
			return nil, status.Errorf(codes.FailedPrecondition, msg)
		}
	} else {
		err := CreateSymlink(device, mountpoint)

		if err != nil {
			msg := fmt.Sprintf("failed to create a link: oldname=%v, newname=%v, %v", device, mountpoint, err)
			glog.Errorf(msg)
			return nil, errors.New(msg)
		}
	}

	glog.V(5).Info("node publish volume success")
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume implementation
func (v *Volume) NodeUnpublishVolume(req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	vol, _, err := v.getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, err
	}

	if CSIFilesystem == vol.Metadata[CSIVolumeMode] {
		// check volume is unmounted
		mounted, err := connector.IsMounted(req.TargetPath)
		if !mounted {
			glog.Info("target path is already unmounted")
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}

		// Umount
		glog.V(5).Infof("mountpoint:%s", req.TargetPath)
		err = connector.Umount(req.TargetPath)
		if err != nil {
			msg := fmt.Sprintf("failed to umount: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	if CSIBlock == vol.Metadata[CSIVolumeMode] {
		_, err = exec.Command("rm", "-rf", req.TargetPath).CombinedOutput()
		if err != nil {
			return nil, err
		}
	}

	glog.V(5).Info("node unpublish volume success")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// getVolumeAndAttachmentByVolumeId Get volume and attachment by volumeId
func (v *Volume) getVolumeAndAttachmentByVolumeId(volId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
	if r := v.getReplicationByVolume(volId); r != nil {
		volId = r.Metadata[AttachedVolumeId]
	}

	vol, err := v.Client.GetVolume(volId)
	if nil != err || nil == vol {
		msg := fmt.Sprintf("volume does not exist: %v", err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := v.Client.ListVolumeAttachments()
	if nil != err {
		msg := fmt.Sprintf("list volume attachments failed: %v", err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.NotFound, msg)
	}

	var attachment *model.VolumeAttachmentSpec

	hostName, err := connector.GetHostName()
	if err != nil {
		msg := fmt.Sprintf("faild to get host name: %v", err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.FailedPrecondition, msg)
	}

	host, err := nbputil.GetHostByHostName(v.Client, hostName)
	if err != nil {
		msg := fmt.Sprintf("faild to get host name: %v", err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.FailedPrecondition, msg)
	}

	for _, attach := range attachments {
		if attach.VolumeId == volId && attach.HostId == host.Id {
			attachment = attach
			break
		}
	}

	if attachment == nil {
		msg := fmt.Sprintf("attachment does not exist")
		glog.Error(msg)
		return nil, nil, status.Error(codes.FailedPrecondition, msg)
	}

	return vol, attachment, nil
}

// getVolumeAndAttachment Get volume and attachment with volumeId and attachmentId
func (v *Volume) getVolumeAndAttachment(volumeId string, attachmentId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
	vol, err := v.Client.GetVolume(volumeId)
	if nil != err || nil == vol {
		msg := fmt.Sprintf("volume %s does not exist: %v", volumeId, err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.NotFound, msg)
	}

	attachment, err := v.Client.GetVolumeAttachment(attachmentId)
	if nil != err || nil == attachment {
		msg := fmt.Sprintf("the volume attachment %s does not exist: %v", attachmentId, err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.FailedPrecondition, msg)
	}

	return vol, attachment, nil
}
