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

// mountDeviceAndUpdateAttachment Mount device and then update attachment
func mountDeviceAndUpdateAttachment(device string, mountpoint string, key string, mountFlags []string, needUpdateAtc bool, attachment *model.VolumeAttachmentSpec, block *csi.VolumeCapability_BlockVolume) error {
	var err error

	if nil == block {
		if len(mountFlags) > 0 {
			_, err = exec.Command("mount", "-o", strings.Join(mountFlags, ","), device, mountpoint).CombinedOutput()
		} else {
			_, err = exec.Command("mount", device, mountpoint).CombinedOutput()
		}

		if nil != err {
			return status.Error(codes.Aborted, fmt.Sprintf("failed to mount: %v", err.Error()))
		}
	} else {
		_, err = os.Lstat(mountpoint)

		if err != nil && os.IsNotExist(err) {
			glog.V(5).Infof("Mountpoint=%v is not exist", mountpoint)
		} else {
			glog.Errorf("Mountpoint=%v already exists!", mountpoint)
			_, err := exec.Command("rm", "-rf", mountpoint).CombinedOutput()

			if nil != err {
				return err
			}
		}

		err = os.Symlink(device, mountpoint)
		if err != nil {
			return err
		}
	}

	// update volume Attachmentment
	paths := strings.Split(attachment.Metadata[key], ";")
	isExist := false
	for _, path := range paths {
		if mountpoint == path {
			isExist = true
			break
		}
	}

	if false == isExist {
		paths = append(paths, mountpoint)
		attachment.Metadata[key] = strings.Join(paths, ";")
		needUpdateAtc = true
	}

	if needUpdateAtc {
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
	isExist := false
	paths := strings.Split(attachment.Metadata[key], ";")
	for index, path := range paths {
		if path == TargetPath {
			modifyPaths = append(paths[:index], paths[index+1:]...)
			isExist = true
			break
		}
	}

	if !isExist {
		return nil
	}

	if (1 == len(modifyPaths) && 0 == len(modifyPaths[0])) || (0 == len(modifyPaths)) {
		glog.V(5).Info("No more " + key)
		delete(attachment.Metadata, key)

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
	} else {
		attachment.Metadata[key] = strings.Join(modifyPaths, ";")
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
	defer glog.V(5).Info("end to NodeStageVolume")

	// Check REQUIRED field
	glog.V(5).Info("start to NodeStageVolume, Volume_id: " + req.VolumeId + ", staging_target_path: " + req.StagingTargetPath)
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

	vol, attachment, err := getVolumeAndAttachment(volId, attachmentId)
	if nil != err {
		return nil, err
	}

	device := attachment.Mountpoint
	mountpoint := req.StagingTargetPath
	needUpdateAtc := false

	if 0 == len(device) || "-" == device {
		volConnector := connector.NewConnector(attachment.DriverVolumeType)
		if nil == volConnector {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("unsupport driverVolumeType: %s", attachment.DriverVolumeType))
		}

		devicePath, err := volConnector.Attach(attachment.ConnectionData)
		if nil != err || 0 == len(devicePath) || "-" == devicePath {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("failed to find device: %s", err.Error()))
		}

		attachment.Mountpoint = devicePath
		device = devicePath
		needUpdateAtc = true
	}

	// Check if it is: "Volume published but is incompatible"
	mnt := req.VolumeCapability.GetMount()
	block := req.VolumeCapability.GetBlock()
	glog.V(5).Infof("VolumeCapability Mount=%+v, Block=%+v\n", mnt, block)

	if (nil != mnt) && (nil != block) {
		return nil, status.Error(codes.InvalidArgument, "volumeMode cannot be both Block and Filesystem")
	}

	if nil == vol.Metadata {
		vol.Metadata = make(map[string]string)
	}

	var mountFlags []string
	if nil == block {
		vol.Metadata[KCSIVolumeMode] = "Filesystem"
		mountFlags = mnt.MountFlags
		_, err = exec.Command("findmnt", device, mountpoint).CombinedOutput()
		glog.V(5).Infof("findmnt err: %v \n", err)

		if nil == err {
			if len(mountFlags) > 0 {
				_, err := exec.Command("findmnt", "-o", strings.Join(mountFlags, ","), device, mountpoint).CombinedOutput()
				if nil != err {
					return nil, status.Error(codes.Aborted, "Volume published but is incompatible")
				}
			}

			return &csi.NodeStageVolumeResponse{}, nil
		}

		// Format
		curFSType := connector.GetFSType(attachment.Mountpoint)
		hopeFSType := DefFSType
		if "" != mnt.FsType {
			hopeFSType = mnt.FsType
		}

		if "" == curFSType {
			_, err := exec.Command("mkfs", "-t", hopeFSType, "-F", device).CombinedOutput()
			if err != nil {
				return nil, status.Error(codes.Aborted, fmt.Sprintf("failed to mkfs: %v", err.Error()))
			}
		} else {
			if "" != mnt.FsType {
				if mnt.FsType != curFSType {
					glog.Errorf("Volume formatted but is incompatible, %v != %v!", mnt.FsType, curFSType)
					return nil, status.Error(codes.Aborted, "Volume formatted but is incompatible")
				}
			}
		}

		// Mount
		_, err = exec.Command("mkdir", "-p", mountpoint).CombinedOutput()
		if err != nil {
			return nil, status.Error(codes.Aborted, fmt.Sprintf("failed to mkdir: %v", err.Error()))
		}
	} else {
		vol.Metadata[KCSIVolumeMode] = "Block"
	}

	err = mountDeviceAndUpdateAttachment(device, mountpoint, KStagingTargetPath, mountFlags, needUpdateAtc, attachment, block)
	if err != nil {
		return nil, err
	}

	vol.Status = model.VolumeInUse
	_, err = Client.UpdateVolume(vol.Id, vol)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "update volume failed")
	}

	glog.V(5).Info("NodeStageVolume success")
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume implementation
func (p *Plugin) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {
	defer glog.V(5).Info("end to NodeUnstageVolume")

	// Check REQUIRED field
	glog.V(5).Info("start to NodeUnstageVolume, Volume_id: " + req.VolumeId + ", staging_target_path: " + req.StagingTargetPath)
	if "" == req.VolumeId || "" == req.StagingTargetPath {
		return nil, status.Error(codes.InvalidArgument, "Volume_id/staging_target_path must be specified")
	}

	vol, attachment, err := getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, err
	}

	if KCSIFilesystem == vol.Metadata[KCSIVolumeMode] {
		err = connector.Umount(req.StagingTargetPath)
		if err != nil {
			return nil, err
		}
	}

	if KCSIBlock == vol.Metadata[KCSIVolumeMode] {
		_, err = exec.Command("rm", "-rf", req.StagingTargetPath).CombinedOutput()
		if err != nil {
			return nil, err
		}
	}

	err = delTargetPathInAttachment(attachment, KStagingTargetPath, req.StagingTargetPath)
	if err != nil {
		return nil, err
	}

	vol.Status = model.VolumeAvailable
	_, err = Client.UpdateVolume(vol.Id, vol)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "update volume failed")
	}

	glog.V(5).Info("NodeUnstageVolume success")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume implementation
func (p *Plugin) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {
	defer glog.V(5).Info("end to NodePublishVolume")

	// Check REQUIRED field
	glog.V(5).Info("start to NodePublishVolume, Volume_id: " + req.VolumeId + ", staging_target_path: " + req.StagingTargetPath + ", target_path: " + req.TargetPath)
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

	// Check if it is: "Volume published but is incompatible"
	mnt := req.VolumeCapability.GetMount()
	block := req.VolumeCapability.GetBlock()
	glog.V(5).Infof("VolumeCapability Mount=%+v, Block=%+v\n", mnt, block)

	if (nil != mnt) && (nil != block) {
		return nil, status.Error(codes.InvalidArgument, "volumeMode cannot be both Block and Filesystem")
	}

	var mountFlags []string
	if nil == block {
		mountFlags = append(mnt.MountFlags, "bind")
		if req.Readonly {
			mountFlags = append(mountFlags, "ro")
		}

		_, err = exec.Command("findmnt", device, mountpoint).CombinedOutput()
		glog.V(5).Infof("findmnt err: %v \n", err)

		if nil == err {
			if len(mountFlags) > 0 {
				_, err := exec.Command("findmnt", "-o", strings.Join(mountFlags, ","), device, mountpoint).CombinedOutput()
				if nil != err {
					return nil, status.Error(codes.Aborted, "Volume published but is incompatible")
				}
			}

			return &csi.NodePublishVolumeResponse{}, nil
		}

	}

	err = mountDeviceAndUpdateAttachment(device, mountpoint, KTargetPath, mountFlags, false, attachment, block)
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
	defer glog.V(5).Info("end to NodeUnpublishVolume")

	// Check REQUIRED field
	glog.V(5).Info("start to NodeUnpublishVolume, Volume_id: " + req.VolumeId + ", target_path: " + req.TargetPath)
	if "" == req.VolumeId || "" == req.TargetPath {
		return nil, status.Error(codes.InvalidArgument, "Volume_id/target_path must be specified")
	}

	vol, attachment, err := getVolumeAndAttachmentByVolumeId(req.VolumeId)
	if err != nil {
		return nil, err
	}

	if KCSIFilesystem == vol.Metadata[KCSIVolumeMode] {
		err := connector.Umount(req.TargetPath)
		if err != nil {
			return nil, err
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

// getNodeId gets node id based on the protocol, i.e., FC, iSCSI, RBD, etc.
func getNodeId() (string, error) {
	hostName, err := connector.GetHostName()
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, err.Error())
	}

	nodeId := hostName
	fcConnector := connector.NewConnector(connector.FcDriver)
	if fcConnector != nil {
		fcInitiator, err := fcConnector.GetInitiatorInfo()
		if err == nil {
			wwpnInterface, ok := fcInitiator.InitiatorData[connector.Wwpn]
			if ok {
				wwpnStrArray, ok := wwpnInterface.([]string)
				if ok {
					for _, wwpnStr := range wwpnStrArray {
						nodeId = nodeId + "," + connector.Wwpn + ":" + wwpnStr
					}
				}
			}

			wwnnInterface, ok := fcInitiator.InitiatorData[connector.Wwnn]
			if ok {
				wwnnStrArray, ok := wwnnInterface.([]string)
				if ok {
					for _, wwnnStr := range wwnnStrArray {
						nodeId = nodeId + "," + connector.Wwnn + ":" + wwnnStr
					}
				}
			}
		}
	}

	iscsiConnector := connector.NewConnector(connector.IscsiDriver)
	if iscsiConnector != nil {
		iscsiInitiator, err := iscsiConnector.GetInitiatorInfo()
		if err == nil {
			iqnInterface, ok := iscsiInitiator.InitiatorData[connector.Iqn]

			if ok {
				iqnStr, ok := iqnInterface.(string)
				if ok {
					nodeId = nodeId + "," + connector.Iqn + ":" + iqnStr
				}
			}
		}
	}

	glog.V(5).Info("NodeId: " + nodeId)
	return nodeId, nil
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

// NodeGetVolumeStats implementation
func (p *Plugin) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
