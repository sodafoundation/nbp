package opensds

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"google.golang.org/grpc/codes"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/client/iscsi"
	sdscontroller "github.com/opensds/nbp/client/opensds"
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
	targetlun := req.PublishVolumeInfo["targetlun"]

	// Connect Target
	log.Printf("[NodePublishVolume] portal:%s targetiqn:%s targetlun:%s volumeid:%s",
		portal, targetiqn, targetlun, req.VolumeId)
	device, err := iscsi.Connect(portal, targetiqn, targetlun)
	if err != nil {
		return nil, err
	}

	// Format and Mount
	log.Printf("[NodePublishVolume] device:%s TargetPath:%s", device, req.TargetPath)
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

	if errCode := p.CheckVersionSupport(req.Version); errCode != codes.OK {
		msg := "the version specified in the request is not supported by the Plugin."
		return nil, status.Error(errCode, msg)
	}

	// Umount
	log.Printf("[NodeUnpublishVolume] TargetPath:%s", req.TargetPath)
	err := iscsi.Umount(req.TargetPath)
	if err != nil {
		return nil, err
	}

	client := sdscontroller.GetClient("")

	//check volume is exist
	volSpec, errVol := client.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", req.VolumeId)
		return nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to NodeUnpublish volume.")
	}

	hostname, _ := os.Hostname()
	for _, attachSpec := range attachments {

		log.Printf("[NodeUnpublishVolume] attachSpec.Host:%s hostname:%s",
			attachSpec.Host, hostname)

		if attachSpec.VolumeId == req.VolumeId && attachSpec.Host == hostname {
			iscsiCon := iscsi.ParseIscsiConnectInfo(attachSpec.ConnectionData)
			// Disconnect
			if iscsiCon != nil {
				err = iscsi.Disconnect(iscsiCon.TgtPortal, iscsiCon.TgtIQN)
				if err != nil {
					return nil, err
				}
			}
		}
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
