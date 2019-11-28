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

package common

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

// ValidateNodeStageVolume - validates input paras of NodeStageVolume request
func ValidateNodeStageVolume(req *csi.NodeStageVolumeRequest) error {

	if "" == req.VolumeId || "" == req.StagingTargetPath || nil == req.VolumeCapability {
		msg := "volume_id/staging_target_path/volume_capability must be specified"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}
	return nil
}

// ValidateNodeUnstageVolume - validates input paras of NodeUnstageVolume request
func ValidateNodeUnstageVolume(req *csi.NodeUnstageVolumeRequest) error {

	if "" == req.VolumeId || "" == req.StagingTargetPath {
		msg := "volume_id/staging_target_path must be specified"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}
	return nil
}

// ValidateNodePublishVolume - validates input paras of NodePublishVolume request
func ValidateNodePublishVolume(req *csi.NodePublishVolumeRequest) error {

	if "" == req.VolumeId || "" == req.StagingTargetPath || "" == req.TargetPath || nil == req.VolumeCapability {
		msg := "volume_id/staging_target_path/target_path/volume_capability must be specified"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}
	return nil
}

// ValidateNodeUnpublishVolume - validates input paras of NodeUnpublishVolume request
func ValidateNodeUnpublishVolume(req *csi.NodeUnpublishVolumeRequest) error {

	if "" == req.VolumeId || "" == req.TargetPath {
		msg := "volume_id/target_path must be specified"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}
	return nil
}

// NodeGetInfo gets information on a node
func NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
	topologyZoneKey string) (
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
				topologyZoneKey: DefaultAvailabilityZone,
			},
		},
	}, nil
}

// NodeGetCapabilities implementation
func NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	glog.V(5).Info("start to node get capabilities")
	defer glog.V(5).Info("end to node get capabilities")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
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
func NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
