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

	"os/exec"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/opensds/nbp/client/iscsi"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/driver"
	c "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Node Service                                    //
////////////////////////////////////////////////////////////////////////////////

var (
	// Client opensds client
	Client *c.Client
)

func init() {
	Client = sdscontroller.GetClient("", "")
}

// GetVolAndAtc implementation
func GetVolAndAtc(volId string, atcId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
	vol, err := Client.GetVolume(volId)
	if nil != err || nil == vol {
		return nil, nil, status.Error(codes.NotFound, "Volume does not exist")
	}

	atc, err := Client.GetVolumeAttachment(atcId)
	if nil != err || nil == atc {
		return nil, nil, status.Error(codes.FailedPrecondition,
			fmt.Sprintf("the volume attachment %s does not exist, ", atcId))
	}

	return vol, atc, nil
}

// MountDeviceAndUpdateAtc implementation
func MountDeviceAndUpdateAtc(device string, mountpoint string, key string, mountFlags []string, needUpdateAtc bool, atc *model.VolumeAttachmentSpec) error {
	var err error

	if len(mountFlags) > 0 {
		_, err = exec.Command("mount", "-o", strings.Join(mountFlags, ","), device, mountpoint).CombinedOutput()
	} else {
		_, err = exec.Command("mount", device, mountpoint).CombinedOutput()
	}

	if nil != err {
		return status.Error(codes.Aborted, fmt.Sprintf("failed to mount: %v", err.Error()))
	}

	// update volume attachment
	paths := strings.Split(atc.Metadata[key], ";")
	isExist := false
	for _, path := range paths {
		if mountpoint == path {
			isExist = true
			break
		}
	}

	if false == isExist {
		paths = append(paths, mountpoint)
		atc.Metadata[key] = strings.Join(paths, ";")
		needUpdateAtc = true
	}

	if needUpdateAtc {
		_, err = Client.UpdateVolumeAttachment(atc.Id, atc)
		if err != nil {
			return status.Error(codes.FailedPrecondition, "update volume attachment failed")
		}
	}

	return nil
}

// GetVolAndAtcWhenUnxx implementation
func GetVolAndAtcWhenUnxx(volId string) (*model.VolumeSpec, *model.VolumeAttachmentSpec, error) {
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

	var atc *model.VolumeAttachmentSpec
	iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == volId && attachSpec.Host == localIqn {
			atc = attachSpec
			break
		}
	}

	return vol, atc, nil
}

// DelTargetPathInAtc implementation
func DelTargetPathInAtc(atc *model.VolumeAttachmentSpec, key string, TargetPath string) error {
	if nil == atc {
		return nil
	}

	if _, exist := atc.Metadata[key]; !exist {
		return nil
	}

	var modifyPaths []string
	isExist := false
	paths := strings.Split(atc.Metadata[key], ";")
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

	if (0 == len(modifyPaths)) && (KStagingTargetPath == key) {
		volDriver := driver.NewVolumeDriver(atc.DriverVolumeType)
		if volDriver == nil {
			return status.Error(codes.FailedPrecondition, fmt.Sprintf("Unsupport driverVolumeType: %s", atc.DriverVolumeType))
		}

		err := volDriver.Detach(atc.ConnectionData)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
		atc.Mountpoint = "-"
	}

	atc.Metadata[key] = strings.Join(modifyPaths, ";")
	_, err := Client.UpdateVolumeAttachment(atc.Id, atc)
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
	attachId := req.PublishInfo[KPublishAttachId]

	if r := getReplicationByVolume(volId); r != nil {
		if r.ReplicationStatus == model.ReplicationFailover {
			volId = r.SecondaryVolumeId
			attachId = req.PublishInfo[KPublishSecondaryAttachId]
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

	vol, atc, err := GetVolAndAtc(volId, attachId)
	if nil != err {
		return nil, err
	}

	device := atc.Mountpoint
	mountpoint := req.StagingTargetPath
	needUpdateAtc := false

	if 0 == len(device) || "-" == device {
		volDriver := driver.NewVolumeDriver(atc.DriverVolumeType)
		if nil == volDriver {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("unsupport driverVolumeType: %s", atc.DriverVolumeType))
		}

		devicePath, err := volDriver.Attach(atc.ConnectionData)
		if nil != err || 0 == len(devicePath) || "-" == devicePath {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("failed to find device: %s", err.Error()))
		}

		atc.Mountpoint = devicePath
		device = devicePath
		needUpdateAtc = true
	}

	// Check if it is: "Volume published but is incompatible"
	mnt := req.VolumeCapability.GetMount()
	mountFlags := mnt.MountFlags
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
	curFSType := iscsi.GetFSType(atc.Mountpoint)
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
				return nil, status.Error(codes.Aborted, "Volume formatted but is incompatible")
			}
		}
	}

	// Mount
	_, err = exec.Command("mkdir", "-p", mountpoint).CombinedOutput()
	if err != nil {
		return nil, status.Error(codes.Aborted, fmt.Sprintf("failed to mkdir: %v", err.Error()))
	}

	err = MountDeviceAndUpdateAtc(device, mountpoint, KStagingTargetPath, mountFlags, needUpdateAtc, atc)
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

	// Umount
	err := iscsi.Umount(req.StagingTargetPath)
	if err != nil {
		return nil, err
	}

	vol, atc, err := GetVolAndAtcWhenUnxx(req.VolumeId)
	if err != nil {
		return nil, err
	}

	err = DelTargetPathInAtc(atc, KStagingTargetPath, req.StagingTargetPath)
	if err != nil {
		return nil, err
	}

	vol.Status = model.VolumeAvailable
	_, err = Client.UpdateVolume(vol.Id, vol)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "update volume failed")
	}

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
	attachId := req.PublishInfo[KPublishAttachId]

	if r := getReplicationByVolume(volId); r != nil {
		volId = r.Metadata[KAttachedVolumeId]
		attachId = r.Metadata[KAttachedId]
	}

	_, atc, err := GetVolAndAtc(volId, attachId)
	if nil != err {
		return nil, err
	}

	device := req.StagingTargetPath
	mountpoint := req.TargetPath
	needUpdateAtc := false

	// Check if it is: "Volume published but is incompatible"
	mnt := req.VolumeCapability.GetMount()
	mountFlags := append(mnt.MountFlags, "bind")
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

	// Mount
	err = MountDeviceAndUpdateAtc(device, mountpoint, KTargetPath, mountFlags, needUpdateAtc, atc)
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

	// Umount
	err := iscsi.Umount(req.TargetPath)
	if err != nil {
		return nil, err
	}

	_, atc, err := GetVolAndAtcWhenUnxx(req.VolumeId)
	if err != nil {
		return nil, err
	}

	err = DelTargetPathInAtc(atc, KTargetPath, req.TargetPath)
	if err != nil {
		return nil, err
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetId implementation
func (p *Plugin) NodeGetId(
	ctx context.Context,
	req *csi.NodeGetIdRequest) (
	*csi.NodeGetIdResponse, error) {

	glog.V(5).Info("start to GetNodeID")
	defer glog.V(5).Info("end to GetNodeID")

	iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}

	return &csi.NodeGetIdResponse{
		NodeId: localIqn,
	}, nil
}

// NodeGetInfo implementation
func (p *Plugin) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
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
