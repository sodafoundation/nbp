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
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

// CreateVolume implementation
func (p *Plugin) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	glog.V(5).Info("start to create volume")
	defer glog.V(5).Info("end to create volume")

	// check input parameters
	if err := common.ValidateCreateVolReq(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.FileShareClient.CreateFileShare(req)
}

// DeleteVolume implementation
func (p *Plugin) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	glog.V(5).Info("start to delete volume")
	defer glog.V(5).Info("end to delete volume")

	// check input parameters
	if err := common.ValidateDelVolReq(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.FileShareClient.DeleteFileShare(req.VolumeId)
}

// ControllerPublishVolume implementation
func (p *Plugin) ControllerPublishVolume(ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	glog.V(5).Info("start to controller publish volume")
	defer glog.V(5).Info("end to controller publish volume")

	// check input parameters
	if err := common.ValidateCtrlPubVolReq(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	glog.V(5).Infof("plugin information %#v", p)

	return p.FileShareClient.ControllerPublishFileShare(req)
}

// ControllerUnpublishVolume implementation
func (p *Plugin) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	glog.V(5).Infof("start to controller unpublish volume")
	defer glog.V(5).Info("end to controller unpublish volume")

	// check input parameters
	if err := common.ValidateCtrlUnPubVolReq(req); err != nil {
		glog.Error(err.Error())
		return nil, err
	}

	return p.FileShareClient.ControllerUnpublishFileShare(req)
}
// ControllerExpandVolume implementation
func (p *Plugin) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest) (
	*csi.ControllerExpandVolumeResponse, error) {

	glog.V(5).Infof("start to controller expand volume")
	defer glog.V(5).Info("end to controller expand volume")

	return nil, status.Error(codes.Unimplemented, "")
}

// ValidateVolumeCapabilities implementation
func (p *Plugin) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	return common.ValidateVolumeCapabilities(ctx, req)
}

// ListVolumes implementation
func (p *Plugin) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	glog.V(5).Info("start to list volumes")
	defer glog.V(5).Info("end to list volumes")

	return p.FileShareClient.ListFileShares(req)
}

// GetCapacity implementation
func (p *Plugin) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	return common.GetCapacity(p.FileShareClient.Client, ctx, req)
}

// ControllerGetCapabilities implementation
func (p *Plugin) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return common.ControllerGetCapabilities(ctx, req)
}

// CreateSnapshot implementation
func (p *Plugin) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {

	return common.CreateSnapshot(p.FileShareClient.Client, ctx, req)
}

// DeleteSnapshot implementation
func (p *Plugin) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {

	return common.DeleteSnapshot(p.FileShareClient.Client, ctx, req)
}

// ListSnapshots implementation
func (p *Plugin) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {

	return common.ListSnapshots(p.FileShareClient.Client, ctx, req)
}
