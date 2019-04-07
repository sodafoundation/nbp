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
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	var err error
	client, err = sdscontroller.GetClient("", "")
	if client == nil || err != nil {
		glog.Errorf("client init failed, %s", err.Error())
		return
	}
}

// getVolumeAndAttachment Get volume and attachment with volumeId and attachmentId
func getVolumeAndAttachment(volumeId string, attachmentId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
	vol, err := client.GetVolume(volumeId)
	if nil != err || nil == vol {
		msg := fmt.Sprintf("volume %s does not exist, %v", volumeId, err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.NotFound, msg)
	}

	attachment, err := client.GetVolumeAttachment(attachmentId)
	if nil != err || nil == attachment {
		msg := fmt.Sprintf("the volume attachment %s does not exist, %v", attachmentId, err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.FailedPrecondition, msg)
	}

	return vol, attachment, nil
}

// updateAttachment Update attachment
func updateAttachment(mountpoint string, key string, attachment *model.VolumeAttachmentSpec) error {
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
		_, err := client.UpdateVolumeAttachment(attachment.Id, attachment)
		if err != nil {
			msg := fmt.Sprintf("update volume attachmentment failed, %v", err)
			glog.Error(msg)
			return status.Error(codes.FailedPrecondition, msg)
		}
	}

	return nil
}

// getVolumeAndAttachmentByVolumeId Get volume and attachment by volumeId
func getVolumeAndAttachmentByVolumeId(volId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
	if r := getReplicationByVolume(volId); r != nil {
		volId = r.Metadata[KAttachedVolumeId]
	}

	vol, err := client.GetVolume(volId)
	if nil != err || nil == vol {
		msg := fmt.Sprintf("Volume does not exist, %v", err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := client.ListVolumeAttachments()
	if nil != err {
		msg := fmt.Sprintf("List volume attachments failed, %v", err)
		glog.Error(msg)
		return nil, nil, status.Error(codes.NotFound, msg)
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

	if attachment == nil {
		msg := fmt.Sprintf("attachment is not exist")
		glog.Error(msg)
		return nil, nil, status.Error(codes.FailedPrecondition, msg)
	}

	return vol, attachment, nil
}

// delTargetPathInAttachment Delete a targetPath (stagingTargetPath) from the attachment
func delTargetPathInAttachment(attachment *model.VolumeAttachmentSpec, key string, TargetPath string) error {
	targetPathList, exist := attachment.Metadata[key]
	if !exist {
		return nil
	}

	paths := strings.Split(targetPathList, ";")
	for index, path := range paths {
		if path == TargetPath {
			paths = append(paths[:index], paths[index+1:]...)
			break
		}
	}

	if 0 == len(paths) {
		glog.V(5).Info("No more " + key)
		delete(attachment.Metadata, key)
	} else {
		attachment.Metadata[key] = strings.Join(paths, ";")
	}

	if KStagingTargetPath == key {
		volConnector := connector.NewConnector(attachment.DriverVolumeType)

		if volConnector == nil {
			msg := fmt.Sprintf("Unsupport driverVolumeType: %s", attachment.DriverVolumeType)
			glog.Error(msg)
			return status.Error(codes.FailedPrecondition, msg)
		}

		err := volConnector.Detach(attachment.ConnectionData)
		if err != nil {
			msg := fmt.Sprintf("detach falied, %v", err)
			glog.Error(msg)
			return status.Errorf(codes.FailedPrecondition, "%s", msg)
		}

		attachment.Mountpoint = "-"
	}

	_, err := client.UpdateVolumeAttachment(attachment.Id, attachment)
	if err != nil {
		msg := "update volume attachment failed"
		glog.Error(msg)
		return status.Error(codes.FailedPrecondition, msg)
	}

	return nil
}

// Symlink implementation
func createSymlink(device, mountpoint string) error {
	_, err := os.Lstat(mountpoint)
	if err != nil && os.IsNotExist(err) {
		glog.V(5).Infof("Mountpoint=%v does not exist", mountpoint)
	} else {
		glog.Errorf("mountpoint=%v already exists", mountpoint)
		// The mountpoint deleted here is a folder or a soft connection.
		// From the test results, this is fine.
		_, err := exec.Command("rm", "-rf", mountpoint).CombinedOutput()

		if nil != err {
			glog.Errorf("Faild to delete %v", mountpoint)
			return err
		}
	}

	err = os.Symlink(device, mountpoint)
	if err != nil {
		glog.Errorf("failed to create a link: oldname=%v, newname=%v\n", device, mountpoint)
		return err
	}

	return nil
}

// NodeStageVolume implementation
func (p *Plugin) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to NodeStageVolume, Volume_id: " + req.VolumeId +
		", staging_target_path: " + req.StagingTargetPath)
	defer glog.V(5).Info("end to NodeStageVolume")

	if "" == req.VolumeId || "" == req.StagingTargetPath || nil == req.VolumeCapability {
		msg := "Volume_id/staging_target_path/volume_capability must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		msg := "NodeStageVolume: the client is nil"
		glog.Info(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
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
		if _, err := client.UpdateReplication(r.Id, r); err != nil {
			msg := fmt.Sprintf("update replication(%s) failed, %v", r.Id, err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	vol, attachment, err := getVolumeAndAttachment(volId, attachmentId)
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
			msg := fmt.Sprintf("failed to find device: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		device = devicePath
	}

	mnt := req.VolumeCapability.GetMount()
	block := req.VolumeCapability.GetBlock()
	glog.V(5).Infof("VolumeCapability Mount=%+v, Block=%+v\n", mnt, block)

	if nil != mnt && nil != block {
		msg := "volumeMode cannot be both Block and Filesystem"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if nil == vol.Metadata {
		vol.Metadata = make(map[string]string)
	}

	if nil == block {
		vol.Metadata[KCSIVolumeMode] = "Filesystem"
		// Format
		hopeFSType := req.PublishContext[KVolumeFstype]
		glog.Infof("fsType is %s", hopeFSType)

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
				msg := fmt.Sprintf("failed to mkfs: %v", err.Error())
				glog.Error(msg)
				return nil, status.Error(codes.Aborted, msg)
			}
		} else {
			glog.Infof("Device: %s has been formatted yet. fsType: %s", device, curFSType)
		}

		// Mount
		mounted, err := connector.IsMounted(mountpoint)
		if err != nil {
			msg := fmt.Sprintf("failed to check mounted, %v", err)
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
			msg := fmt.Sprintf("failed to mount, %v", err)
			glog.Errorf(msg)
			return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
		}
	} else {
		vol.Metadata[KCSIVolumeMode] = "Block"
		err = createSymlink(device, mountpoint)

		if err != nil {
			msg := fmt.Sprintf("failed to create a link: oldname=%v, newname=%v, err %v", device, mountpoint, err)
			glog.Error(msg)
			return nil, status.Error(codes.Aborted, msg)
		}
	}

	err = updateAttachment(mountpoint, KStagingTargetPath, attachment)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	_, err = client.UpdateVolume(vol.Id, vol)
	if err != nil {
		msg := fmt.Sprintf("update volume failed, %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
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
	glog.V(5).Info("start to NodeUnstageVolume, Volume_id: " + req.VolumeId +
		", staging_target_path: " + req.StagingTargetPath)
	defer glog.V(5).Info("end to NodeUnstageVolume")

	if "" == req.VolumeId || "" == req.StagingTargetPath {
		msg := "Volume_id/staging_target_path must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		msg := "client is nil"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	vol, attachment, err := getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, err
	}

	if KCSIFilesystem == vol.Metadata[KCSIVolumeMode] {
		// check volume is unmounted
		mounted, err := connector.IsMounted(req.StagingTargetPath)
		if err != nil {
			msg := fmt.Sprintf("check volume is unmounted failed, %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		if !mounted {
			glog.Info("target path is already unmounted")
			return &csi.NodeUnstageVolumeResponse{}, nil
		}

		// Umount
		glog.Infof("[NodeUnpublishVolume] mountpoint:%s", req.StagingTargetPath)
		err = connector.Umount(req.StagingTargetPath)
		if err != nil {
			msg := fmt.Sprintf("failed to umount, %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	if KCSIBlock == vol.Metadata[KCSIVolumeMode] {
		_, err = exec.Command("rm", "-rf", req.StagingTargetPath).CombinedOutput()
		if err != nil {
			msg := fmt.Sprintf("rm -rf %v failed, return error:%v", req.StagingTargetPath, err)
			glog.Errorf(msg)
			return nil, errors.New(msg)
		}
	}

	err = delTargetPathInAttachment(attachment, KStagingTargetPath, req.StagingTargetPath)
	if err != nil {
		return nil, err
	}

	glog.V(5).Info("NodeUnstageVolume success")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume implementation
func (p *Plugin) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to NodePublishVolume, Volume_id: " + req.VolumeId +
		", staging_target_path: " + req.StagingTargetPath + ", target_path: " + req.TargetPath)
	defer glog.V(5).Info("end to NodePublishVolume")

	if "" == req.VolumeId || "" == req.StagingTargetPath || "" == req.TargetPath || nil == req.VolumeCapability {
		msg := "Volume_id/staging_target_path/target_path/volume_capability must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		msg := "client is nil"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
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
	block := req.VolumeCapability.GetBlock()
	glog.V(5).Infof("VolumeCapability Mount=%+v, Block=%+v\n", mnt, block)

	if nil != mnt && nil != block {
		msg := "volumeMode cannot be both Block and Filesystem"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if nil == block {
		// Bind mount
		mountFlags := append(mnt.MountFlags, "bind")
		glog.Infof("req.Readonly, %v", req.Readonly)
		if req.Readonly {
			mountFlags = append(mountFlags, "ro")
		}

		fsType := req.PublishContext[KVolumeFstype]
		glog.Infof("fsType is %s", fsType)
		if mnt.FsType != "" {
			glog.Infof("use system fsType %s", mnt.FsType)
			fsType = mnt.FsType
		}

		// Mount
		mounted, err := connector.IsMounted(mountpoint)
		if err != nil {
			msg := fmt.Sprintf("failed to check mounted, %v", err)
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
			msg := fmt.Sprintf("failed to mount, %v", err)
			glog.Errorf(msg)
			return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
		}
	} else {
		err = createSymlink(device, mountpoint)

		if err != nil {
			msg := fmt.Sprintf("failed to create a link: oldname=%v, newname=%v, %v", device, mountpoint, err)
			glog.Errorf(msg)
			return nil, errors.New(msg)
		}
	}

	// update volume attachment
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
		msg := "Volume_id/target_path must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		msg := "client is nil"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	vol, attachment, err := getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, err
	}

	if KCSIFilesystem == vol.Metadata[KCSIVolumeMode] {
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
			msg := fmt.Sprintf("failed to Umount, %v", err)
			glog.Info(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	if KCSIBlock == vol.Metadata[KCSIVolumeMode] {
		_, err = exec.Command("rm", "-rf", req.TargetPath).CombinedOutput()
		if err != nil {
			return nil, err
		}
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

	if client == nil {
		msg := "client is nil"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	hostName, err := connector.GetHostName()
	if err != nil {
		msg := fmt.Sprintf("failed to get node name %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	var initiators []string

	volDriverTypes := []string{connector.FcDriver, connector.IscsiDriver}

	for _, volDriverType := range volDriverTypes {
		volDriver := connector.NewConnector(volDriverType)
		if volDriver == nil {
			glog.Errorf("Unsupport volDriver: %s", volDriverType)
			continue
		}

		initiator, err := volDriver.GetInitiatorInfo()
		if err != nil {
			glog.Errorf("Cannot get initiator for driver volume type %s, err: %v", volDriverType, err)
			continue
		}

		initiators = append(initiators, initiator)
	}

	if len(initiators) == 0 {
		msg := fmt.Sprintf("Cannot get any initiator for host %s", hostName)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	nodeId := hostName + "," + strings.Join(initiators, ",")

	glog.Infof("node info is %s", nodeId)

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

// NodeGetVolumeStats implementation
func (p *Plugin) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
