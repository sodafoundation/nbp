package opensds

import (
	"log"
	"runtime"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
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
	}
	if req.Parameters != nil && req.Parameters["AvailabilityZone"] != "" {
		volumebody.AvailabilityZone = req.Parameters["AvailabilityZone"]
	}

	v, err := c.CreateVolume(volumebody)
	if err != nil {
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
	// TODO
	return nil, nil
}

// ControllerUnpublishVolume implementation
func (p *Plugin) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {
	// TODO
	return nil, nil
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
