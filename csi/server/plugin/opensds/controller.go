package opensds

import (
	"log"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
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
		Handle: &csi.VolumeHandle{
			Id: v.Id,
			Metadata: map[string]string{
				"Name":             v.Name,
				"Status":           v.Status,
				"AvailabilityZone": v.AvailabilityZone,
				"PoolId":           v.PoolId,
				"ProfileId":        v.ProfileId,
			},
		},
		CapacityBytes: uint64(v.Size),
	}

	return &csi.CreateVolumeResponse{
		Reply: &csi.CreateVolumeResponse_Result_{
			Result: &csi.CreateVolumeResponse_Result{
				VolumeInfo: volumeinfo,
			},
		},
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
	err := c.DeleteVolume(req.VolumeHandle.Id, &model.VolumeSpec{})
	if err != nil {
		return nil, err
	}

	return &csi.DeleteVolumeResponse{
		Reply: &csi.DeleteVolumeResponse_Result_{
			Result: &csi.DeleteVolumeResponse_Result{},
		},
	}, nil
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

	volumeid := req.VolumeInfo.Handle.Id
	if strings.TrimSpace(volumeid) == "" {
		return &csi.ValidateVolumeCapabilitiesResponse{
			Reply: &csi.ValidateVolumeCapabilitiesResponse_Error{
				Error: &csi.Error{
					Value: &csi.Error_ValidateVolumeCapabilitiesError_{
						ValidateVolumeCapabilitiesError: &csi.Error_ValidateVolumeCapabilitiesError{
							ErrorCode:        csi.Error_ValidateVolumeCapabilitiesError_INVALID_VOLUME_INFO,
							ErrorDescription: "invalid volume id",
						},
					},
				},
			},
		}, nil
	}

	for _, capabilities := range req.VolumeCapabilities {
		if capabilities.GetMount() != nil {
			return &csi.ValidateVolumeCapabilitiesResponse{
				Reply: &csi.ValidateVolumeCapabilitiesResponse_Result_{
					Result: &csi.ValidateVolumeCapabilitiesResponse_Result{
						Supported: false,
						Message:   "opensds does not support mounted volume",
					},
				},
			}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Result_{
			Result: &csi.ValidateVolumeCapabilitiesResponse_Result{
				Supported: true,
				Message:   "supported",
			},
		},
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

	ens := []*csi.ListVolumesResponse_Result_Entry{}
	for _, v := range volumes {
		if v != nil {
			volumeinfo := &csi.VolumeInfo{
				Handle: &csi.VolumeHandle{
					Id: v.Id,
					Metadata: map[string]string{
						"Name":             v.Name,
						"Status":           v.Status,
						"AvailabilityZone": v.AvailabilityZone,
						"PoolId":           v.PoolId,
						"ProfileId":        v.ProfileId,
					},
				},
				CapacityBytes: uint64(v.Size),
			}

			ens = append(ens, &csi.ListVolumesResponse_Result_Entry{
				VolumeInfo: volumeinfo,
			})
		}
	}

	return &csi.ListVolumesResponse{
		Reply: &csi.ListVolumesResponse_Result_{
			Result: &csi.ListVolumesResponse_Result{
				Entries: ens,
			},
		},
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
		Reply: &csi.GetCapacityResponse_Result_{
			Result: &csi.GetCapacityResponse_Result{
				AvailableCapacity: freecapacity,
			},
		},
	}, nil
}

// ControllerGetCapabilities implementation
func (p *Plugin) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	log.Println("start to ControllerGetCapabilities")
	defer log.Println("end to ControllerGetCapabilities")

	return &csi.ControllerGetCapabilitiesResponse{
		Reply: &csi.ControllerGetCapabilitiesResponse_Result_{
			Result: &csi.ControllerGetCapabilitiesResponse_Result{
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
			},
		},
	}, nil
}
