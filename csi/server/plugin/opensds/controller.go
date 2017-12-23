package opensds

import (
	"log"
	"runtime"
	"strings"

	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/client/iscsi"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/opensds/pkg/model"
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

	log.Println("start to CreateVolume")
	defer log.Println("end to CreateVolume")

	c := sdscontroller.GetClient("")

	// build volume body
	volumebody := &model.VolumeSpec{}
	volumebody.Name = req.Name
	if req.CapacityRange != nil {
		volumebody.Size = int64(req.CapacityRange.RequiredBytes)
	} else {
		//Using default volume size
		volumebody.Size = 1
	}
	if req.Parameters != nil && req.Parameters["AvailabilityZone"] != "" {
		volumebody.AvailabilityZone = req.Parameters["AvailabilityZone"]
	}

	v, err := c.CreateVolume(volumebody)
	if err != nil {
		log.Fatalf("failed to CreateVolume: %v", err)
		return nil, err
	}

	// return volume info
	volumeinfo := &csi.VolumeInfo{
		CapacityBytes: uint64(v.Size),
		Id:            v.Id,
		Attributes: map[string]string{
			"Name":             v.Name,
			"Status":           v.Status,
			"AvailabilityZone": v.AvailabilityZone,
			"PoolId":           v.PoolId,
			"ProfileId":        v.ProfileId,
			"lvPath":           v.Metadata["lvPath"],
		},
	}

	return &csi.CreateVolumeResponse{
		VolumeInfo: volumeinfo,
	}, nil
}

// DeleteVolume implementation
func (p *Plugin) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	log.Println("start to DeleteVolume")
	defer log.Println("end to DeleteVolume")

	c := sdscontroller.GetClient("")
	err := c.DeleteVolume(req.VolumeId, &model.VolumeSpec{})
	if err != nil {
		return nil, err
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume implementation
func (p *Plugin) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	log.Println("start to ControllerPublishVolume")
	defer log.Println("end to ControllerPublishVolume")

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

	//TODO: need to check if node exists?

	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish volume.")
	}

	var attachNodes []string
	hostname := req.NodeId
	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == req.VolumeId && attachSpec.Host != hostname {
			//TODO: node id is what? use hostname to indicate node id currently.
			attachNodes = append(attachNodes, attachSpec.Host)
		}
	}

	if len(attachNodes) != 0 {
		//if the volume has been published, but without MULTI_NODE capability, return error.
		mode := req.VolumeCapability.AccessMode.Mode
		if mode != csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER {
			msg := fmt.Sprintf("the volume %s has been published to another node.", req.VolumeId)
			return nil, status.Error(codes.AlreadyExists, msg)
		}
	}

	iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}

	attachReq := &model.VolumeAttachmentSpec{
		VolumeId: req.VolumeId,
		HostInfo: model.HostInfo{
			Host:      req.NodeId,
			Platform:  runtime.GOARCH,
			OsType:    runtime.GOOS,
			Ip:        iscsi.GetHostIp(),
			Initiator: localIqn,
		},
		Metadata: req.VolumeAttributes,
	}
	attachSpec, errAttach := client.CreateVolumeAttachment(attachReq)
	if errAttach != nil {
		msg := fmt.Sprintf("the volume %s failed to publish to node %s.", req.VolumeId, req.NodeId)
		log.Fatalf("failed to ControllerPublishVolume: %v", attachReq)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	return &csi.ControllerPublishVolumeResponse{
		PublishVolumeInfo: map[string]string{
			"ip":     attachSpec.Ip,
			"host":   attachSpec.Host,
			"atcid":  attachSpec.Id,
			"status": attachSpec.Status,
		},
	}, nil
}

// ControllerUnpublishVolume implementation
func (p *Plugin) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	log.Println("start to ControllerUnpublishVolume")
	defer log.Println("end to ControllerUnpublishVolume")

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
		return nil, status.Error(codes.FailedPrecondition, "Failed to unpublish volume.")
	}

	var acts []*model.VolumeAttachmentSpec
	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == req.VolumeId && (req.NodeId == "" || attachSpec.Host == req.NodeId) {
			acts = append(acts, attachSpec)
		}
	}

	for _, act := range acts {
		err = client.DeleteVolumeAttachment(act.Id, act)
		if err != nil {
			msg := fmt.Sprintf("the volume %s failed to unpublish from node %s.", req.VolumeId, req.NodeId)
			log.Fatalf("failed to ControllerUnpublishVolume: %v", err)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities implementation
func (p *Plugin) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	log.Println("start to ValidateVolumeCapabilities")
	defer log.Println("end to ValidateVolumeCapabilities")

	if strings.TrimSpace(req.VolumeId) == "" {
		// csi.Error_ValidateVolumeCapabilitiesError_INVALID_VOLUME_INFO
		return nil, status.Error(codes.NotFound, "invalid volume id")
	}

	for _, capabilities := range req.VolumeCapabilities {
		if capabilities.GetMount() != nil {
			return &csi.ValidateVolumeCapabilitiesResponse{
				Supported: false,
				Message:   "opensds does not support mounted volume",
			}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Supported: true,
		Message:   "supported",
	}, nil
}

// ListVolumes implementation
func (p *Plugin) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	log.Println("start to ListVolumes")
	defer log.Println("end to ListVolumes")

	c := sdscontroller.GetClient("")

	// only support list all the volumes at present
	volumes, err := c.ListVolumes()
	if err != nil {
		return nil, err
	}

	ens := []*csi.ListVolumesResponse_Entry{}
	for _, v := range volumes {
		if v != nil {

			volumeinfo := &csi.VolumeInfo{
				CapacityBytes: uint64(v.Size),
				Id:            v.Id,
				Attributes: map[string]string{
					"Name":             v.Name,
					"Status":           v.Status,
					"AvailabilityZone": v.AvailabilityZone,
					"PoolId":           v.PoolId,
					"ProfileId":        v.ProfileId,
				},
			}

			ens = append(ens, &csi.ListVolumesResponse_Entry{
				VolumeInfo: volumeinfo,
			})
		}
	}

	return &csi.ListVolumesResponse{
		Entries: ens,
	}, nil
}

// GetCapacity implementation
func (p *Plugin) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	log.Println("start to GetCapacity")
	defer log.Println("end to GetCapacity")

	c := sdscontroller.GetClient("")

	pools, err := c.ListPools()
	if err != nil {
		return nil, err
	}

	// calculate all the free capacity of pools
	freecapacity := uint64(0)
	for _, p := range pools {
		if p != nil {
			freecapacity += uint64(p.FreeCapacity)
		}
	}

	return &csi.GetCapacityResponse{
		AvailableCapacity: freecapacity,
	}, nil
}

// ControllerProbe implementation
func (p *Plugin) ControllerProbe(
	ctx context.Context,
	req *csi.ControllerProbeRequest) (
	*csi.ControllerProbeResponse, error) {

	log.Println("start to ControllerProbe")
	defer log.Println("end to ControllerProbe")

	switch runtime.GOOS {
	case "linux":
		return &csi.ControllerProbeResponse{}, nil
	default:
		msg := "unsupported operating system:" + runtime.GOOS
		log.Fatalf(msg)
		// csi.Error_ControllerProbeError_MISSING_REQUIRED_HOST_DEPENDENCY
		return nil, status.Error(codes.FailedPrecondition, msg)
	}
}

// ControllerGetCapabilities implementation
func (p *Plugin) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	log.Println("start to ControllerGetCapabilities")
	defer log.Println("end to ControllerGetCapabilities")

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			&csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			&csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
			&csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
					},
				},
			},
			&csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_GET_CAPACITY,
					},
				},
			},
		},
	}, nil
}
