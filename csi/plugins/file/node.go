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

package file

import (
	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/opensds/nbp/csi/common"
	"golang.org/x/net/context"
        "google.golang.org/grpc/codes"
        "google.golang.org/grpc/status"
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

	// check input parameters
	if err := common.ValidateNodeUnstageVolume(req); err != nil {
		glog.Error(err.Error())
		return nil, err
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

	// check input parameters
	if err := common.ValidateNodePublishVolume(req); err != nil {
		glog.Error(err.Error())
		return nil, err
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

	// check input parameters
	if err := common.ValidateNodeUnpublishVolume(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.FileShareClient.NodeUnpublishFileShare(req)
}

// NodeExpandVolume implementation
func (p *Plugin) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {

	glog.V(5).Info("start to node expand volume")

	defer glog.V(5).Info("end to node expand volume")
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetInfo gets information on a node
func (p *Plugin) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {

	return common.NodeGetInfo(ctx, req, TopologyZoneKey, p.FileShareClient.Client)
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
