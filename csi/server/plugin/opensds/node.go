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
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"
	"strings"
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

	if errCode := p.CheckVersionSupport(req.Version); errCode != codes.OK {
		msg := "the version specified in the request is not supported by the Plugin."
		return nil, status.Error(errCode, msg)
	}

	client := sdscontroller.GetClient("")

	//check volume is exist
	volSpec, errVol := client.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", req.VolumeId)
		return nil, status.Error(codes.NotFound, msg)
	}

	atc, atcErr := client.GetVolumeAttachment(req.PublishVolumeInfo["atcid"])
	if atcErr != nil || atc == nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
	}

	var targetPaths []string
	if tps, exist := atc.Metadata["target_path"]; exist && len(tps) != 0 {
		targetPaths = strings.Split(tps, ";")
		for _, tp := range targetPaths {
			if req.TargetPath == tp {
				return &csi.NodePublishVolumeResponse{}, nil
			}
		}

		// if volume don't have MULTI_NODE capability, just termination.
		mode := req.VolumeCapability.AccessMode.Mode
		if mode != csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER {
			msg := fmt.Sprintf("the volume %s has been published to this node.", req.VolumeId)
			return nil, status.Error(codes.Aborted, msg)
		}
	}

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

	// obtain attachments to decide if can format.
	atcs, err := client.ListVolumeAttachments()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
	}
	format := true
	for _, attachSpec := range atcs {
		if attachSpec.VolumeId == req.VolumeId {
			if _, exist := attachSpec.Metadata["target_path"]; exist {
				// The device is formatted, can't be reformat for shared storage.
				format = false
				break
			}
		}
	}

	// Format and Mount
	log.Printf("[NodePublishVolume] device:%s TargetPath:%s", device, req.TargetPath)
	if format {
		err = iscsi.FormatandMount(device, "", req.TargetPath)
	} else {
		err = iscsi.Mount(device, req.TargetPath)
	}
	if err != nil {
		return nil, err
	}

	targetPaths = append(targetPaths, req.TargetPath)
	atc.Metadata["target_path"] = strings.Join(targetPaths, ";")
	_, err = client.UpdateVolumeAttachment(atc.Id, atc)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
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

	var atc *model.VolumeAttachmentSpec
	hostname, _ := os.Hostname()
	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == req.VolumeId && attachSpec.Host == hostname {
			atc = attachSpec
			break
		}
	}

	if atc == nil {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	if _, exist := atc.Metadata["target_path"]; !exist {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	var modifyTargetPaths []string
	tpExist := false
	targetPaths := strings.Split(atc.Metadata["target_path"], ";")
	for index, path := range targetPaths {
		if path == req.TargetPath {
			modifyTargetPaths = append(targetPaths[:index], targetPaths[index+1:]...)
			tpExist = true
			break
		}
	}
	if !tpExist {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// Umount
	log.Printf("[NodeUnpublishVolume] TargetPath:%s", req.TargetPath)
	err = iscsi.Umount(req.TargetPath)
	if err != nil {
		return nil, err
	}

	if len(modifyTargetPaths) == 0 {
		iscsiCon := iscsi.ParseIscsiConnectInfo(atc.ConnectionData)
		// Disconnect
		if iscsiCon != nil {
			err = iscsi.Disconnect(iscsiCon.TgtPortal, iscsiCon.TgtIQN)
			if err != nil {
				return nil, err
			}
		}
	}

	atc.Metadata["target_path"] = strings.Join(modifyTargetPaths, ";")
	_, err = client.UpdateVolumeAttachment(atc.Id, atc)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to NodeUnpublish volume.")
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
