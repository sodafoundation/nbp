package opensds

import (
	"log"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
	"fmt"
	client2 "github.com/opensds/opensds/client"
)

var (
	VOLUME_ATTACH_NODES_MAX = 100
)

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

func newGeneralError(
	errCode csi.Error_GeneralError_GeneralErrorCode,
	retry bool,
	message string) *csi.ControllerPublishVolumeResponse {
	return &csi.ControllerPublishVolumeResponse{
		Reply: &csi.ControllerPublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:          errCode,
						CallerMustNotRetry: retry,
						ErrorDescription:   message,
					},
				},
			},
		},
	}
}

func newControllerPublishVolumeResponseError(
	errCode csi.Error_ControllerPublishVolumeError_ControllerPublishVolumeErrorCode,
	nodeIds []*csi.NodeID,
	message string) *csi.ControllerPublishVolumeResponse {
	return &csi.ControllerPublishVolumeResponse{
		Reply: &csi.ControllerPublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_ControllerPublishVolumeError_{
					ControllerPublishVolumeError: &csi.Error_ControllerPublishVolumeError{
						ErrorCode:          errCode,
						ErrorDescription:   message,
						NodeIds: nodeIds,
					},
				},
			},
		},
	}
}

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

	log.Println("start to ControllerPublishVolume")
	defer log.Println("end to ControllerPublishVolume")

	if support, errCode := p.CheckVersionSupport(req.Version); !support {
		retry := false
		if errCode == csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION {
			//if version not supported, caller should not try again.
			retry = true
		}
		return newGeneralError(errCode, retry, "The request version is not supported."), nil
	}

	client := sdscontroller.GetClient("")

	//check volume is exist
	volSpec, errVol := client.GetVolume(req.VolumeHandle.Id)
	if errVol != nil || volSpec == nil {
		return newControllerPublishVolumeResponseError(csi.Error_ControllerPublishVolumeError_VOLUME_DOES_NOT_EXIST,
			nil, "the volume is not exist."), nil
	}

	//need to check node exist?

	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return newGeneralError(csi.Error_GeneralError_UNDEFINED, false, "Internal error."), nil
	}

	var attachNodes []*csi.NodeID
	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == req.VolumeHandle.Id {
			node := &csi.NodeID{
				Values: map[string]string{
					"ip": attachSpec.HostInfo.Ip,
					"host": attachSpec.HostInfo.Host,
				},
			}
			attachNodes = append(attachNodes, node)
		}
	}

	if len(attachNodes) != 0 {
		if len(attachNodes) >= VOLUME_ATTACH_NODES_MAX {
			return newControllerPublishVolumeResponseError(csi.Error_ControllerPublishVolumeError_MAX_ATTACHED_NODES,
				attachNodes, "the node to attath has reach max."), nil
		}

		//if the volume has been published, but without MULTI_NODE capability, return error.
		mode := req.VolumeCapability.AccessMode.Mode
		if mode != csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER {
			return newControllerPublishVolumeResponseError(csi.Error_ControllerPublishVolumeError_VOLUME_ALREADY_PUBLISHED,
				attachNodes, "the volume has been published to another node."), nil
		}
	}

	attachReq := &model.VolumeAttachmentSpec{
		VolumeId: req.VolumeHandle.Id,
		HostInfo: &model.HostInfo{
			Ip: req.NodeId.Values["ip"],
			Host: req.NodeId.Values["host"],
		},
	}
	attachSpec, errAttach := client.CreateVolumeAttachment(client2.VolumeAttachmentBuilder(attachReq))
	if errAttach != nil {
		msg := fmt.Sprintf("the volume %s failed to attach to node %s.", req.VolumeHandle.Id, req.NodeId.Values["host"])
		return newControllerPublishVolumeResponseError(csi.Error_ControllerPublishVolumeError_VOLUME_ALREADY_PUBLISHED,
			attachNodes, msg), nil
	}

	return &csi.ControllerPublishVolumeResponse{
		Reply: &csi.ControllerPublishVolumeResponse_Result_{
			Result: &csi.ControllerPublishVolumeResponse_Result{
				PublishVolumeInfo: map[string]string{
					"ip": attachSpec.Ip,
					"host": attachSpec.Host,
					"attachid": attachSpec.Id,
					"status": attachSpec.Status,
				},
			},
		},
	}, nil
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
