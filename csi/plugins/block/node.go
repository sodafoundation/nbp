// Copyright 2018 The OpenSDS Authors.
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
	"fmt"
	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/opensds/nbp/csi/common"
	"github.com/opensds/nbp/csi/util"
	"github.com/opensds/opensds/contrib/connector"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"strings"
)

////////////////////////////////////////////////////////////////////////////////
//                            Node Service                                    //
////////////////////////////////////////////////////////////////////////////////

// NodeStageVolume implementation
func (p *Plugin) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to node stage volume, Volume_id: " + req.VolumeId +
		", staging_target_path: " + req.StagingTargetPath)
	defer glog.V(5).Info("end to node stage volume")

	// check input parameters
	if err := common.ValidateNodeStageVolume(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.VolumeClient.NodeStageVolume(req)
}

// NodeUnstageVolume implementation
func (p *Plugin) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to node unstage volume, volume_id: " + req.VolumeId +
		", staging_target_path: " + req.StagingTargetPath)
	defer glog.V(5).Info("end to node unstage volume")

	// check input parameters
	if err := common.ValidateNodeUnstageVolume(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.VolumeClient.NodeUnstageVolume(req)
}

// NodePublishVolume implementation
func (p *Plugin) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to node publish volume, volume_id: " + req.VolumeId +
		", staging_target_path: " + req.StagingTargetPath + ", target_path: " + req.TargetPath)
	defer glog.V(5).Info("end to node publish volume")

	// check input parameters
	if err := common.ValidateNodePublishVolume(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.VolumeClient.NodePublishVolume(req)
}

// NodeUnpublishVolume implementation
func (p *Plugin) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to node unpublish volume, volume_id: " + req.VolumeId + ", target_path: " + req.TargetPath)
	defer glog.V(5).Info("end to node unpublish volume")

	// check input parameters
	if err := common.ValidateNodeUnpublishVolume(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.VolumeClient.NodeUnpublishVolume(req)
}

// NodeExpandVolume implementation
func (p *Plugin) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {

	glog.V(5).Info("start to node expand volume")

	// Check REQUIRED field
	defer glog.V(5).Info("end to node expand volume")
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	capacityBytes := common.GetSize(req.GetCapacityRange())
	volSizeBytes := capacityBytes * util.GiB

	args := []string{"-o", "source", "--noheadings", "--target", req.GetVolumePath()}
	output, err := connector.ExecCmd("findmnt", args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine device path: %v", err)
	}

	devicePath := strings.TrimSpace(string(output))
	if len(devicePath) == 0 {
		return nil, status.Errorf(codes.Internal, "Could not get valid device for mount path: %q", req.GetVolumePath())
	}

	if _, err := resize(devicePath, req.GetVolumePath()); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not resize volume %q (%q):  %v", volumeID, devicePath, err)
	}

	glog.V(5).Infof("Node expand volume completed for  volumeId: %v,  size: %v", volumeID, volSizeBytes)
	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: volSizeBytes,
	}, nil
}

// NodeGetInfo gets information on a node
func (p *Plugin) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {

	return common.NodeGetInfo(ctx, req, TopologyZoneKey, p.VolumeClient.Client)
}

// NodeGetCapabilities implementation
func (p *Plugin) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return common.NodeGetCapabilities(ctx, req)
}

// NodeGetVolumeStats implementation
func (p *Plugin) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {

	return common.NodeGetVolumeStats(ctx, req)
}

// Resize perform resize of file system
func resize(devicePath string, deviceMountPath string) (bool, error) {
	format, err := connector.GetFSType(devicePath)

	if err != nil {
		formatErr := fmt.Errorf("ResizeFS.Resize - error checking format for device %s: %v", devicePath, err)
		return false, formatErr
	}

	// If disk has no format, there is no need to resize the disk because mkfs.*
	// by default will use whole disk anyways.
	if format == "" {
		return false, nil
	}

	glog.V(5).Infof("File system type for the device: %v", format)
	switch format {
	case "ext3", "ext4":
		return extResize(devicePath)
	case "xfs":
		return xfsResize(deviceMountPath)
	}
	return false, fmt.Errorf("Resize of format %s is not supported for device %s mounted at %s", format, devicePath, deviceMountPath)
}

func extResize(devicePath string) (bool, error) {

	Para := fmt.Sprintf("lsscsi | grep %s", devicePath)
	output, err := connector.ExecCmd("/bin/bash", "-c", Para)
	if err != nil {
		glog.V(5).Infof("Failed to execute lsscsi for device output : %v, err : %v", output, err)
		glog.Error(err.Error())
		return false, err
	}

	//Output looks like : [4:0:0:1]    disk    IET      VIRTUAL-DISK     0001  /dev/sdb
	// Need to parse and get host identifier (4 in above case)
	glog.V(5).Infof("end to node expand volume, lsscsi: %v", output)
	hostId := strings.Split(output, " ")[0]
	hostId = strings.Split(hostId, ":")[0]
	hostId = strings.Split(hostId, "[")[1]

	scanCommandPara := "'- - -' > /sys/class/scsi_host/host" + hostId + "/scan"

	output, err = connector.ExecCmd("echo", scanCommandPara)
	if err != nil {
		glog.V(5).Infof("Failed to execute scan command for device output : %v, err : %v", output, err)
		glog.Error(err.Error())
		return false, err
	}

	deviceIndex := strings.Split(devicePath, "/dev/")[1]

	reScanCommandPara := "'1' > /sys/block/" + deviceIndex + "/device/rescan"
	output, err = connector.ExecCmd("echo", reScanCommandPara)
	if err != nil {
		glog.V(5).Infof("Failed to execute rescan command for device output : %v, err : %v", output, err)
		glog.Error(err.Error())
		return false, err
	}

	output, err = connector.ExecCmd("resize2fs", devicePath)
	if err != nil {
		glog.V(5).Infof("Failed to execute resize command for device path : %v", devicePath)
		glog.Error(err.Error())
		return false, err
	}

	glog.V(5).Infof("Resize success for device path : %v", devicePath)
	return true, nil

}

func xfsResize(deviceMountPath string) (bool, error) {
	args := []string{"-d", deviceMountPath}
	output, err := connector.ExecCmd("xfs_growfs", args...)

	if err == nil {
		return true, nil
	}

	resizeError := fmt.Errorf("resize of device %s failed: %v. xfs_growfs output: %s", deviceMountPath, err, string(output))
	return false, resizeError
}
