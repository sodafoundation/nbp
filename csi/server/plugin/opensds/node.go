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

package opensds

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/opensds/opensds/contrib/connector"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Node Service                                    //
////////////////////////////////////////////////////////////////////////////////

// Symlink implementation
func createSymlink(device, mountpoint string) error {
	_, err := os.Lstat(mountpoint)
	if err != nil && os.IsNotExist(err) {
		glog.V(5).Infof("mountpoint=%v does not exist", mountpoint)
	} else {
		glog.Errorf("mountpoint=%v already exists", mountpoint)
		// The mountpoint deleted here is a folder or a soft connection.
		// From the test results, this is fine.
		_, err := exec.Command("rm", "-rf", mountpoint).CombinedOutput()

		if nil != err {
			glog.Errorf("faild to delete %v", mountpoint)
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
	glog.V(5).Info("start to node stage volume, Volume_id: " + req.VolumeId +
		", staging_target_path: " + req.StagingTargetPath)
	defer glog.V(5).Info("end to node stage volume")

	if "" == req.VolumeId || "" == req.StagingTargetPath || nil == req.VolumeCapability {
		msg := "volume_id/staging_target_path/volume_capability must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if p.PluginStorageType == VolumeStorageType {
		return p.VolumeClient.NodeStageVolume(req)
	}

	return p.FileShareClient.NodeStageFileShare(req)
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

	if "" == req.VolumeId || "" == req.StagingTargetPath {
		msg := "volume_id/staging_target_path must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if p.PluginStorageType == VolumeStorageType {
		return p.VolumeClient.NodeUnstageVolume(req)
	}

	return p.FileShareClient.NodeUnstageFileShare(req)
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

	if "" == req.VolumeId || "" == req.StagingTargetPath || "" == req.TargetPath || nil == req.VolumeCapability {
		msg := "volume_id/staging_target_path/target_path/volume_capability must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if p.PluginStorageType == VolumeStorageType {
		return p.VolumeClient.NodePublishVolume(req)
	}

	return p.FileShareClient.NodePublishFileShare(req)
}

// NodeUnpublishVolume implementation
func (p *Plugin) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	// Check REQUIRED field
	glog.V(5).Info("start to node unpublish volume, volume_id: " + req.VolumeId + ", target_path: " + req.TargetPath)
	defer glog.V(5).Info("end to node unpublish volume")

	if "" == req.VolumeId || "" == req.TargetPath {
		msg := "volume_id/target_path must be specified"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if p.PluginStorageType == VolumeStorageType {
		return p.VolumeClient.NodeUnpublishVolume(req)
	}

	return p.FileShareClient.NodeUnpublishFileShare(req)
}

// NodeGetInfo gets information on a node
func (p *Plugin) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {

	glog.Info("start to get node info")
	defer glog.Info("end to get node info")

	hostName, err := connector.GetHostName()
	if err != nil {
		msg := fmt.Sprintf("failed to get node name: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	var initiators []string

	volDriverTypes := []string{connector.FcDriver, connector.IscsiDriver, connector.NvmeofDriver}

	for _, volDriverType := range volDriverTypes {
		volDriver := connector.NewConnector(volDriverType)
		if volDriver == nil {
			glog.Errorf("unsupport volume driver: %s", volDriverType)
			continue
		}

		initiator, err := volDriver.GetInitiatorInfo()
		if err != nil {
			glog.Errorf("cannot get initiator for driver volume type %s, err: %v", volDriverType, err)
			continue
		}

		initiators = append(initiators, initiator)
	}

	if len(initiators) == 0 {
		msg := fmt.Sprintf("cannot get any initiator for host %s", hostName)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	nodeId := hostName + "," + strings.Join(initiators, ",") + "," + connector.GetHostIP()

	glog.Infof("node info is %s", nodeId)

	return &csi.NodeGetInfoResponse{
		NodeId: nodeId,
		// driver works only on this zone
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				TopologyZoneKey: DefaultAvailabilityZone,
			},
		},
	}, nil
}

// NodeGetCapabilities implementation
func (p *Plugin) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	glog.V(5).Info("start to node get capabilities")
	defer glog.V(5).Info("end to node get capabilities")

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
