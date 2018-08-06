// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
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

package opensds

import (
	"fmt"
	"runtime"
	"strings"


	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/csi/util"
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

	glog.Info("start to CreateVolume")
	defer glog.Info("end to CreateVolume")

	c := sdscontroller.GetClient("", "")

	// build volume body
	volumebody := &model.VolumeSpec{}
	volumebody.Name = req.Name
	if req.CapacityRange != nil {
		volumeSizeBytes := int64(req.CapacityRange.RequiredBytes)
		allocationUnitBytes := int64(1024 * 1024 * 1024)
		volumebody.Size = (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
		if volumebody.Size < 1 {
			//Using default volume size
			volumebody.Size = 1
		}
	} else {
		//Using default volume size
		volumebody.Size = 1
	}
	var secondaryAZ = util.OpensdsDefaultSecondaryAZ
	var enableReplication = false
	for k, v := range req.GetParameters() {
		switch strings.ToLower(k) {
		case KParamProfile:
			volumebody.ProfileId = v
		case KParamAZ:
			volumebody.AvailabilityZone = v
		case KParamEnableReplication:
			if strings.ToLower(v) == "true" {
				enableReplication = true
			}
		case KParamSecondaryAZ:
			secondaryAZ = v
		}
	}

	glog.Infof("CreateVolume volumebody: %v", volumebody)

	v, err := c.CreateVolume(volumebody)
	if err != nil {
		glog.Fatalf("failed to CreateVolume: %v", err)
		return nil, err
	}

	// return volume info
	volumeinfo := &csi.Volume{
		CapacityBytes: v.Size,
		Id:            v.Id,
		Attributes: map[string]string{
			KVolumeName:      v.Name,
			KVolumeStatus:    v.Status,
			KVolumeAZ:        v.AvailabilityZone,
			KVolumePoolId:    v.PoolId,
			KVolumeProfileId: v.ProfileId,
			KVolumeLvPath:    v.Metadata["lvPath"],
		},
	}

	if enableReplication {
		volumebody.AvailabilityZone = secondaryAZ
		volumebody.Name = SecondaryPrefix + req.Name
		sVol, err := c.CreateVolume(volumebody)
		if err != nil {
			glog.Errorf("failed to create secondar volume: %v", err)
			return nil, err
		}
		replicaBody := &model.ReplicationSpec{
			Name:              req.Name,
			PrimaryVolumeId:   v.Id,
			SecondaryVolumeId: sVol.Id,
			ReplicationMode:   model.ReplicationModeSync,
			ReplicationPeriod: 0,
		}
		replicaResp, err := c.CreateReplication(replicaBody)
		if err != nil {
			glog.Errorf("Create replication failed,:%v", err)
			return nil, err
		}
		volumeinfo.Attributes[KVolumeReplicationId] = replicaResp.Id
	}

	return &csi.CreateVolumeResponse{
		Volume: volumeinfo,
	}, nil
}

func getReplicationByVolume(volId string) *model.ReplicationSpec {
	c := sdscontroller.GetClient("","")
	replications, _ := c.ListReplications()
	for _, r := range replications {
		if volId == r.PrimaryVolumeId || volId == r.SecondaryVolumeId {
			return r
		}
	}
	return nil
}

// DeleteVolume implementation
func (p *Plugin) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {
	glog.Info("start to DeleteVolume")
	defer glog.Info("end to DeleteVolume")
	volId := req.VolumeId
	c := sdscontroller.GetClient("", "")
	r := getReplicationByVolume(volId)
	if r != nil {
		if err := c.DeleteReplication(r.Id, nil); err != nil {
			return nil, err
		}
		if err := c.DeleteVolume(r.PrimaryVolumeId, &model.VolumeSpec{}); err != nil {
			return nil, err
		}
		if err := c.DeleteVolume(r.SecondaryVolumeId, &model.VolumeSpec{}); err != nil {
			return nil, err
		}
	} else {
		if err := c.DeleteVolume(volId, &model.VolumeSpec{}); err != nil {
			return nil, err
		}
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume implementation
func (p *Plugin) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	glog.Info("start to ControllerPublishVolume")
	defer glog.Info("end to ControllerPublishVolume")

	client := sdscontroller.GetClient("", "")

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

	/*iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}*/
	//NodeId is Node Iqn
	localIqn := req.NodeId

	attachReq := &model.VolumeAttachmentSpec{
		VolumeId: req.VolumeId,
		HostInfo: model.HostInfo{
			Host:      req.NodeId,
			Platform:  runtime.GOARCH,
			OsType:    runtime.GOOS,
			Initiator: localIqn,
		},
		Metadata: req.VolumeAttributes,
	}
	attachSpec, errAttach := client.CreateVolumeAttachment(attachReq)
	if errAttach != nil {
		msg := fmt.Sprintf("the volume %s failed to publish to node %s.", req.VolumeId, req.NodeId)
		glog.Errorf("failed to ControllerPublishVolume: %v", attachReq)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	resp := &csi.ControllerPublishVolumeResponse{
		PublishInfo: map[string]string{
			KPublishHostIp:       attachSpec.Ip,
			KPublishHostName:     attachSpec.Host,
			KPublishAttachId:     attachSpec.Id,
			KPublishAttachStatus: attachSpec.Status,
		},
	}
	if replicationId, ok := req.VolumeAttributes[KVolumeReplicationId]; ok {
		r, err := client.GetReplication(replicationId)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, "Get replication failed")
		}
		attachReq.VolumeId = r.SecondaryVolumeId
		attachSpec, errAttach := client.CreateVolumeAttachment(attachReq)
		if errAttach != nil {
			msg := fmt.Sprintf("the volume %s failed to publish to node %s.", req.VolumeId, req.NodeId)
			glog.Errorf("failed to ControllerPublishVolume: %v", attachReq)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
		resp.PublishInfo[KPublishSecondaryAttachId] = attachSpec.Id
	}
	return resp, nil
}

// ControllerUnpublishVolume implementation
func (p *Plugin) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	glog.Info("start to ControllerUnpublishVolume")
	defer glog.Info("end to ControllerUnpublishVolume")

	client := sdscontroller.GetClient("", "")

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

	if r := getReplicationByVolume(req.VolumeId); r != nil {
		for _, attachSpec := range attachments {
			if attachSpec.VolumeId == r.SecondaryVolumeId && (req.NodeId == "" || attachSpec.Host == req.NodeId) {
				acts = append(acts, attachSpec)
			}
		}
	}
	for _, act := range acts {
		err = client.DeleteVolumeAttachment(act.Id, act)
		if err != nil {
			msg := fmt.Sprintf("the volume %s failed to unpublish from node %s.", req.VolumeId, req.NodeId)
			glog.Errorf("failed to ControllerUnpublishVolume: %v", err)
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

	glog.Info("start to ValidateVolumeCapabilities")
	defer glog.Info("end to ValidateVolumeCapabilities")

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

	glog.Info("start to ListVolumes")
	defer glog.Info("end to ListVolumes")

	c := sdscontroller.GetClient("", "")

	// only support list all the volumes at present
	volumes, err := c.ListVolumes()
	if err != nil {
		return nil, err
	}

	ens := []*csi.ListVolumesResponse_Entry{}
	for _, v := range volumes {
		if v != nil {

			volumeinfo := &csi.Volume{
				CapacityBytes: v.Size,
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
				Volume: volumeinfo,
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

	glog.Info("start to GetCapacity")
	defer glog.Info("end to GetCapacity")

	c := sdscontroller.GetClient("", "")

	pools, err := c.ListPools()
	if err != nil {
		return nil, err
	}

	// calculate all the free capacity of pools
	freecapacity := int64(0)
	for _, p := range pools {
		if p != nil {
			freecapacity += int64(p.FreeCapacity)
		}
	}

	return &csi.GetCapacityResponse{
		AvailableCapacity: freecapacity,
	}, nil
}

// ControllerGetCapabilities implementation
func (p *Plugin) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	glog.Info("start to ControllerGetCapabilities")
	defer glog.Info("end to ControllerGetCapabilities")

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
