package main

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/csi/server/plugin"
	"golang.org/x/net/context"
)

type server struct {
	plugin plugin.Service
}

////////////////////////////////////////////////////////////////////////////////
//                            Identity Service                                //
////////////////////////////////////////////////////////////////////////////////

// GetSupportedVersions
func (s *server) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {
	// Use plugin implementation
	return s.plugin.GetSupportedVersions(ctx, req)
}

// GetPluginInfo
func (s *server) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {
	// Use plugin implementation
	return s.plugin.GetPluginInfo(ctx, req)
}

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

// CreateVolume
func (s *server) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {
	// Use plugin implementation
	return s.plugin.CreateVolume(ctx, req)
}

// DeleteVolume
func (s *server) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {
	// Use plugin implementation
	return s.plugin.DeleteVolume(ctx, req)
}

// ControllerPublishVolume implementation
func (s *server) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {
	// Use plugin implementation
	return s.plugin.ControllerPublishVolume(ctx, req)
}

// ControllerUnpublishVolume implementation
func (s *server) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {
	// Use plugin implementation
	return s.plugin.ControllerUnpublishVolume(ctx, req)
}

// ValidateVolumeCapabilities implementation
func (s *server) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	// Use plugin implementation
	return s.plugin.ValidateVolumeCapabilities(ctx, req)
}

// ListVolumes implementation
func (s *server) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {
	// Use plugin implementation
	return s.plugin.ListVolumes(ctx, req)
}

// GetCapacity implementation
func (s *server) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {
	// Use plugin implementation
	return s.plugin.GetCapacity(ctx, req)
}

// ControllerProbe implementation
func (s *server) ControllerProbe(
	ctx context.Context,
	req *csi.ControllerProbeRequest) (
	*csi.ControllerProbeResponse, error) {
	// Use plugin implementation
	return s.plugin.ControllerProbe(ctx, req)
}

// ControllerGetCapabilities implementation
func (s *server) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {
	// Use plugin implementation
	return s.plugin.ControllerGetCapabilities(ctx, req)
}

////////////////////////////////////////////////////////////////////////////////
//                            Node Service                                    //
////////////////////////////////////////////////////////////////////////////////

// NodePublishVolume implementation
func (s *server) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {
	// Use plugin implementation
	return s.plugin.NodePublishVolume(ctx, req)
}

// NodeUnpublishVolume implementation
func (s *server) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {
	// Use plugin implementation
	return s.plugin.NodeUnpublishVolume(ctx, req)
}

// GetNodeID implementation
func (s *server) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {
	// Use plugin implementation
	return s.plugin.GetNodeID(ctx, req)
}

// NodeProbe implementation
func (s *server) NodeProbe(
	ctx context.Context,
	req *csi.NodeProbeRequest) (
	*csi.NodeProbeResponse, error) {
	// Use plugin implementation
	return s.plugin.NodeProbe(ctx, req)
}

// NodeGetCapabilities implementation
func (s *server) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {
	// Use plugin implementation
	return s.plugin.NodeGetCapabilities(ctx, req)
}
