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
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	k8util "github.com/kubernetes-incubator/external-storage/lib/util"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/csi/util"
	c "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	"github.com/opensds/opensds/pkg/utils/constants"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

var (
	// Client opensds client
	Client *c.Client
)

func init() {
	Client = sdscontroller.GetClient("", "")
}

// CreateVolume implementation
func (p *Plugin) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	glog.V(5).Info("start to CreateVolume")
	defer glog.V(5).Info("end to CreateVolume")

	// build volume body
	volumebody := &model.VolumeSpec{}
	volumebody.Name = req.Name
	allocationUnitBytes := int64(1024 * 1024 * 1024)
	if req.CapacityRange != nil {
		volumeSizeBytes := int64(req.CapacityRange.RequiredBytes)
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

	glog.V(5).Infof("CreateVolume volumebody: %v", volumebody)

	v, err := Client.CreateVolume(volumebody)
	if err != nil {
		glog.Fatalf("failed to CreateVolume: %v", err)
		return nil, err
	}

	// return volume info
	volumeinfo := &csi.Volume{
		CapacityBytes: v.Size * allocationUnitBytes,
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
		sVol, err := Client.CreateVolume(volumebody)
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
		replicaResp, err := Client.CreateReplication(replicaBody)
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
	replications, _ := Client.ListReplications()
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
	glog.V(5).Info("start to DeleteVolume")
	defer glog.V(5).Info("end to DeleteVolume")
	volId := req.VolumeId

	r := getReplicationByVolume(volId)
	if r != nil {
		if err := Client.DeleteReplication(r.Id, nil); err != nil {
			return nil, err
		}
		if err := Client.DeleteVolume(r.PrimaryVolumeId, &model.VolumeSpec{}); err != nil {
			return nil, err
		}
		if err := Client.DeleteVolume(r.SecondaryVolumeId, &model.VolumeSpec{}); err != nil {
			return nil, err
		}
	} else {
		if err := Client.DeleteVolume(volId, &model.VolumeSpec{}); err != nil {
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

	glog.V(5).Info("start to ControllerPublishVolume")
	defer glog.V(5).Info("end to ControllerPublishVolume")

	//check volume is exist
	volSpec, errVol := Client.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", req.VolumeId)
		return nil, status.Error(codes.NotFound, msg)
	}

	//TODO: need to check if node exists?

	attachments, err := Client.ListVolumeAttachments()
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
	attachSpec, errAttach := Client.CreateVolumeAttachment(attachReq)
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
		r, err := Client.GetReplication(replicationId)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, "Get replication failed")
		}
		attachReq.VolumeId = r.SecondaryVolumeId
		attachSpec, errAttach := Client.CreateVolumeAttachment(attachReq)
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

	glog.V(5).Info("start to ControllerUnpublishVolume")
	defer glog.V(5).Info("end to ControllerUnpublishVolume")

	//check volume is exist
	volSpec, errVol := Client.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", req.VolumeId)
		return nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := Client.ListVolumeAttachments()
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
		err = Client.DeleteVolumeAttachment(act.Id, act)
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

	glog.V(5).Info("start to ValidateVolumeCapabilities")
	defer glog.V(5).Info("end to ValidateVolumeCapabilities")

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

	glog.V(5).Info("start to ListVolumes")
	defer glog.V(5).Info("end to ListVolumes")

	// only support list all the volumes at present
	volumes, err := Client.ListVolumes()
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

	glog.V(5).Info("start to GetCapacity")
	defer glog.V(5).Info("end to GetCapacity")

	pools, err := Client.ListPools()
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

	glog.V(5).Info("start to ControllerGetCapabilities")
	defer glog.V(5).Info("end to ControllerGetCapabilities")

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
			&csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
					},
				},
			},
			&csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
					},
				},
			},
		},
	}, nil
}

// CreateSnapshot implementation
func (p *Plugin) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {

	defer glog.V(5).Info("end to CreateSnapshot")
	glog.V(5).Infof("start to CreateSnapshot, Name: %v, SourceVolumeId: %v, CreateSnapshotSecrets: %v, parameters: %v!",
		req.Name, req.SourceVolumeId, req.CreateSnapshotSecrets, req.Parameters)

	if 0 == len(req.Name) {
		return nil, status.Error(codes.InvalidArgument, "Snapshot Name cannot be empty")
	}

	if 0 == len(req.SourceVolumeId) {
		return nil, status.Error(codes.InvalidArgument, "Source Volume ID cannot be empty")
	}

	snapReq := &model.VolumeSnapshotSpec{
		Name:     req.Name,
		VolumeId: req.SourceVolumeId,
	}

	snapshot, err := Client.CreateVolumeSnapshot(snapReq)
	if nil != err {
		return nil, err
	}

	createdAt, err := time.Parse(constants.TimeFormat, snapshot.CreatedAt)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      snapshot.Size * k8util.GiB,
			Id:             snapshot.Id,
			SourceVolumeId: snapshot.VolumeId,
			CreatedAt:      createdAt.UnixNano(),
			Status: &csi.SnapshotStatus{
				Type: csi.SnapshotStatus_READY,
			},
		},
	}, nil
}

// DeleteSnapshot implementation
func (p *Plugin) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {

	defer glog.V(5).Info("end to DeleteSnapshot")
	glog.V(5).Infof("start to DeleteSnapshot, SnapshotId: %v, DeleteSnapshotSecrets: %v!",
		req.SnapshotId, req.DeleteSnapshotSecrets)

	if 0 == len(req.SnapshotId) {
		return nil, status.Error(codes.InvalidArgument, "Snapshot ID cannot be empty")
	}

	err := Client.DeleteVolumeSnapshot(req.SnapshotId, nil)

	if nil != err {
		return nil, err
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots implementation
func (p *Plugin) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {

	defer glog.V(5).Info("end to ListSnapshots")
	glog.V(5).Infof("start to ListSnapshots, MaxEntries: %v, StartingToken: %v, SourceVolumeId: %v, SnapshotId: %v!",
		req.MaxEntries, req.StartingToken, req.SourceVolumeId, req.SnapshotId)

	var opts = map[string]string{"Id": req.SnapshotId, "VolumeId": req.SourceVolumeId}
	snapshots, err := Client.ListVolumeSnapshots(opts)

	if nil != err {
		return nil, err
	}

	entries := []*csi.ListSnapshotsResponse_Entry{}
	for _, snapshot := range snapshots {
		createdAt, err := time.Parse(constants.TimeFormat, snapshot.CreatedAt)
		if nil != err {
			return nil, status.Error(codes.Internal, err.Error())
		}

		entries = append(entries, &csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SizeBytes:      snapshot.Size * k8util.GiB,
				Id:             snapshot.Id,
				SourceVolumeId: snapshot.VolumeId,
				CreatedAt:      createdAt.UnixNano(),
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		})
	}

	return &csi.ListSnapshotsResponse{Entries: entries}, nil
}
