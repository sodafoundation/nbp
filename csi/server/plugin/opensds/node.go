package opensds

import (
	"log"
	"os"
	"runtime"

	"google.golang.org/grpc/codes"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/client/iscsi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Node Service                                    //
////////////////////////////////////////////////////////////////////////////////

// NodePublishVolume implementation
func (p *Plugin) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	log.Println("start to NodePublishVolume")
	defer log.Println("end to NodePublishVolume")

	portal := req.PublishVolumeInfo["portal"]
	targetiqn := req.PublishVolumeInfo["targetiqn"]
	targetlun := req.VolumeId

	// Connect Target
	device, err := iscsi.Connect(portal, targetiqn, targetlun)
	if err != nil {
		return nil, err
	}

	// Format and Mount
	err = iscsi.FormatandMount(device, "", req.TargetPath)
	if err != nil {
		return nil, err
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume implementation
func (p *Plugin) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	log.Println("start to NodeUnpublishVolume")
	defer log.Println("end to NodeUnpublishVolume")

	// Umount
	err := iscsi.Umount(req.TargetPath)
	if err != nil {
		return nil, err
	}

	// Disconnect
	// TODO: get portal and targetiqn
	err = iscsi.Disconnect("", "")
	if err != nil {
		return nil, err
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// GetNodeID implementation
func (p *Plugin) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	log.Println("start to GetNodeID")
	defer log.Println("end to GetNodeID")

	// Get host name from os
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &csi.GetNodeIDResponse{
		NodeId: hostname,
	}, nil
}

// NodeProbe implementation
func (p *Plugin) NodeProbe(
	ctx context.Context,
	req *csi.NodeProbeRequest) (
	*csi.NodeProbeResponse, error) {

	log.Println("start to NodeProbe")
	defer log.Println("end to NodeProbe")

	switch runtime.GOOS {
	case "linux":
		return &csi.NodeProbeResponse{}, nil
	default:
		msg := "unsupported operating system:" + runtime.GOOS
		log.Fatalf(msg)
		// csi.Error_NodeProbeError_MISSING_REQUIRED_HOST_DEPENDENCY
		return nil, status.Error(codes.FailedPrecondition, msg)
	}
}

// NodeGetCapabilities implementation
func (p *Plugin) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	log.Println("start to NodeGetCapabilities")
	defer log.Println("end to NodeGetCapabilities")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			&csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}
