// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
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

package opensds

import (
	"fmt"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/opensds/contrib/connector"
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Node Service                                    //
////////////////////////////////////////////////////////////////////////////////

func init() {
	Client = sdscontroller.GetClient("", "")
}

// getVolumeAndAttachment Get volume and attachment with volumeId and attachmentId
func getVolumeAndAttachment(volumeId string, attachmentId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
	vol, err := Client.GetVolume(volumeId)
	if nil != err || nil == vol {
		return nil, nil, status.Error(codes.NotFound, "Volume does not exist")
	}

	attachment, err := Client.GetVolumeAttachment(attachmentId)
	if nil != err || nil == attachment {
		return nil, nil, status.Error(codes.FailedPrecondition,
			fmt.Sprintf("the volume attachment %s does not exist", attachmentId))
	}

	return vol, attachment, nil
}

// updateAttachment Update attachment
func updateAttachment(mountpoint string, key string, attachment *model.VolumeAttachmentSpec) error {
	var err error

	// update volume Attachmentment
	paths := strings.Split(attachment.Metadata[key], ";")
	isExist := false

	for _, path := range paths {
		if mountpoint == path {
			isExist = true
			break
		}
	}

	if !isExist {
		paths = append(paths, mountpoint)
		attachment.Metadata[key] = strings.Join(paths, ";")
		_, err = Client.UpdateVolumeAttachment(attachment.Id, attachment)
		if err != nil {
			return status.Error(codes.FailedPrecondition, "update volume attachmentment failed")
		}
	}

	return nil
}

// getVolumeAndAttachmentByVolumeId Get volume and attachment with volumeId
func getVolumeAndAttachmentByVolumeId(volId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
	if r := getReplicationByVolume(volId); r != nil {
		volId = r.Metadata[KAttachedVolumeId]
	}

	vol, err := Client.GetVolume(volId)
	if nil != err || nil == vol {
		return nil, nil, status.Error(codes.NotFound, "Volume does not exist")
	}

	attachments, err := Client.ListVolumeAttachments()
	if nil != err {
		return nil, nil, status.Error(codes.FailedPrecondition, "List volume attachments failed")
	}

	var attachment *model.VolumeAttachmentSpec
	hostName, err := connector.GetHostName()

	if err != nil {
		msg := fmt.Sprintf("Faild to get host name %v", err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.FailedPrecondition, msg)
	}

	for _, attach := range attachments {
		if attach.VolumeId == volId && attach.Host == hostName {
			attachment = attach
			break
		}
	}

	return vol, attachment, nil
}

// delTargetPathInAttachment Delete a targetPath (stagingTargetPath) from the attachment
func delTargetPathInAttachment(attachment *model.VolumeAttachmentSpec, key string, TargetPath string) error {
	if nil == attachment {
		return nil
	}

	if _, exist := attachment.Metadata[key]; !exist {
		return nil
	}

	var modifyPaths []string
	paths := strings.Split(attachment.Metadata[key], ";")
	for index, path := range paths {
		if path == TargetPath {
			modifyPaths = append(paths[:index], paths[index+1:]...)
			break
		}
	}

	if (1 == len(modifyPaths) && 0 == len(modifyPaths[0])) || (0 == len(modifyPaths)) {
		glog.V(5).Info("No more " + key)
		delete(attachment.Metadata, key)
	} else {
		attachment.Metadata[key] = strings.Join(modifyPaths, ";")
	}

	if KStagingTargetPath == key {
		volConnector := connector.NewConnector(attachment.DriverVolumeType)

		if volConnector == nil {
			return status.Error(codes.FailedPrecondition, fmt.Sprintf("Unsupport driverVolumeType: %s", attachment.DriverVolumeType))
		}

		err := volConnector.Detach(attachment.ConnectionData)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}

		attachment.Mountpoint = "-"
	}

	_, err := Client.UpdateVolumeAttachment(attachment.Id, attachment)
	if err != nil {
		return status.Error(codes.FailedPrecondition, "update volume attachment failed")
	}

	return nil
}

// NodeStageVolume implementation
func (p *Plugin) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to NodeStageVolume, Volume_id: " + req.VolumeId + ", staging_target_path: " + req.StagingTargetPath)
	defer glog.V(5).Info("end to NodeStageVolume")

	if "" == req.VolumeId || "" == req.StagingTargetPath || nil == req.VolumeCapability {
		return nil, status.Error(codes.InvalidArgument, "Volume_id/staging_target_path/volume_capability must be specified")
	}

	volId := req.VolumeId
	attachmentId := req.PublishContext[KPublishAttachId]

	if r := getReplicationByVolume(volId); r != nil {
		if r.ReplicationStatus == model.ReplicationFailover {
			volId = r.SecondaryVolumeId
			attachmentId = req.PublishContext[KPublishSecondaryAttachId]
		}
		if r.Metadata == nil {
			r.Metadata = make(map[string]string)
		}
		r.Metadata[KAttachedVolumeId] = volId
		if _, err := Client.UpdateReplication(r.Id, r); err != nil {
			msg := fmt.Sprintf("update replication(%s) failed, %v", r.Id, err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	_, attachment, err := getVolumeAndAttachment(volId, attachmentId)
	if nil != err {
		return nil, err
	}

	device := attachment.Mountpoint
	mountpoint := req.StagingTargetPath

	if 0 == len(device) || "-" == device {
		volConnector := connector.NewConnector(attachment.DriverVolumeType)
		if nil == volConnector {
			msg := fmt.Sprintf("unsupport driverVolumeType: %s", attachment.DriverVolumeType)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		devicePath, err := volConnector.Attach(attachment.ConnectionData)
		if nil != err || 0 == len(devicePath) || "-" == devicePath {
			msg := fmt.Sprintf("failed to find device: %s", err.Error())
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		device = devicePath
	}

	mnt := req.VolumeCapability.GetMount()
	mountFlags := mnt.MountFlags

	// Format
	hopeFSType := req.PublishContext[KVolumeFstype]
	fmt.Println("fsType is ", hopeFSType)

	if mnt.FsType != "" {
		hopeFSType = mnt.FsType
	}

	curFSType, err := connector.GetFSType(device)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	if curFSType == "" {
		if err := connector.Format(device, hopeFSType); err != nil {
			return nil, status.Error(codes.Aborted, fmt.Sprintf("failed to mkfs: %v", err.Error()))
		}
	} else {
		glog.Infof("Device: %s has been formatted yet. fsType: %s", device, curFSType)
	}

	// Mount
	mounted, err := connector.IsMounted(mountpoint)
	if err != nil {
		msg := fmt.Sprintf("Failed to check mounted, %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
	}

	if mounted {
		glog.Info("volume is already mounted.")
		return &csi.NodeStageVolumeResponse{}, nil
	}

	glog.Info("mounting...")

	err = connector.Mount(device, mountpoint, hopeFSType, mountFlags)
	if err != nil {
		msg := fmt.Sprintf("Failed to mount, %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
	}

	err = updateAttachment(mountpoint, KStagingTargetPath, attachment)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	glog.V(5).Info("NodeStageVolume success")
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume implementation
func (p *Plugin) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to NodeUnstageVolume, Volume_id: " + req.VolumeId + ", staging_target_path: " + req.StagingTargetPath)
	defer glog.V(5).Info("end to NodeUnstageVolume")

	if "" == req.VolumeId || "" == req.StagingTargetPath {
		return nil, status.Error(codes.InvalidArgument, "Volume_id/staging_target_path must be specified")
	}

	_, attachment, err := getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, err
	}

	//check volume is unmounted
	mounted, err := connector.IsMounted(req.StagingTargetPath)
	if !mounted {
		glog.Info("target path is already unmounted")
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	// Umount
	glog.Infof("[NodeUnpublishVolume] mountpoint:%s", req.StagingTargetPath)
	err = connector.Umount(req.StagingTargetPath)
	if err != nil {
		msg := fmt.Sprintf("Failed to Umount, %v", err)
		glog.Info(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	err = delTargetPathInAttachment(attachment, KStagingTargetPath, req.StagingTargetPath)
	if err != nil {
		return nil, err
	}

	//	vol.Status = model.VolumeAvailable
	//	_, err = client.UpdateVolume(vol.Id, vol)
	//	if err != nil {
	//		return nil, status.Error(codes.FailedPrecondition, "update volume failed")
	//	}

	glog.V(5).Info("NodeUnstageVolume success")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume implementation
func (p *Plugin) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to NodePublishVolume, Volume_id: " + req.VolumeId + ", staging_target_path: " + req.StagingTargetPath + ", target_path: " + req.TargetPath)
	defer glog.V(5).Info("end to NodePublishVolume")

	if "" == req.VolumeId || "" == req.StagingTargetPath || "" == req.TargetPath || nil == req.VolumeCapability {
		return nil, status.Error(codes.InvalidArgument, "Volume_id/staging_target_path/target_path/volume_capability must be specified")
	}

	volId := req.VolumeId
	attachmentId := req.PublishContext[KPublishAttachId]

	if r := getReplicationByVolume(volId); r != nil {
		volId = r.Metadata[KAttachedVolumeId]
		attachmentId = r.Metadata[KAttachedId]
	}

	_, attachment, err := getVolumeAndAttachment(volId, attachmentId)
	if nil != err {
		return nil, err
	}

	device := req.StagingTargetPath
	mountpoint := req.TargetPath

	mnt := req.VolumeCapability.GetMount()
	mountFlags := mnt.MountFlags

	// Bind mount
	mountFlags = append(mountFlags, "bind")
	fmt.Println("req.Readonly", req.Readonly)
	if req.Readonly {
		mountFlags = append(mountFlags, "ro")
	}

	fsType := req.PublishContext[KVolumeFstype]
	fmt.Println("fsType is ", fsType)
	if mnt.FsType != "" {
		fsType = mnt.FsType
	}

	// Mount
	mounted, err := connector.IsMounted(mountpoint)
	if err != nil {
		msg := fmt.Sprintf("Failed to check mounted, %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
	}

	if mounted {
		glog.Info("volume is already mounted.")
		return &csi.NodePublishVolumeResponse{}, nil
	}

	glog.Info("mounting...")

	err = connector.Mount(device, mountpoint, fsType, mountFlags)
	if err != nil {
		msg := fmt.Sprintf("Failed to mount, %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
	}

	// Mount
	err = updateAttachment(mountpoint, KTargetPath, attachment)
	if err != nil {
		return nil, err
	}

	glog.V(5).Info("NodePublishVolume success")
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume implementation
func (p *Plugin) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to NodeUnpublishVolume, Volume_id: " + req.VolumeId + ", target_path: " + req.TargetPath)
	defer glog.V(5).Info("end to NodeUnpublishVolume")

	if "" == req.VolumeId || "" == req.TargetPath {
		return nil, status.Error(codes.InvalidArgument, "Volume_id/target_path must be specified")
	}

	// check volume is unmounted
	mounted, err := connector.IsMounted(req.TargetPath)
	if !mounted {
		glog.Info("target path is already unmounted")
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// Umount
	glog.Infof("[NodeUnpublishVolume] mountpoint:%s", req.TargetPath)
	err = connector.Umount(req.TargetPath)
	if err != nil {
		msg := fmt.Sprintf("Failed to Umount, %v", err)
		glog.Info(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	_, attachment, err := getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, err
	}

	err = delTargetPathInAttachment(attachment, KTargetPath, req.TargetPath)
	if err != nil {
		return nil, err
	}

	glog.V(5).Info("NodeUnpublishVolume success")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetInfo gets information on a node
func (p *Plugin) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {
	glog.Info("start to GetNodeInfo")
	defer glog.Info("end to GetNodeInfo")

	nodeId, err := getNodeId()
	if err != nil {
		return nil, err
	}

	return &csi.NodeGetInfoResponse{
		NodeId: nodeId,
	}, nil
}

// NodeGetCapabilities implementation
func (p *Plugin) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	glog.V(5).Info("start to NodeGetCapabilities")
	defer glog.V(5).Info("end to NodeGetCapabilities")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			&csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}

// NodeGetVolumeStats
func (p *Plugin) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
