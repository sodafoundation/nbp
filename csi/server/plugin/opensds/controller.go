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
	"container/list"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/csi/util"
	c "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/contrib/connector"
	"github.com/opensds/opensds/pkg/model"
	"github.com/opensds/opensds/pkg/utils"
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
	client *c.Client
)

func init() {
	var err error
	client, err = sdscontroller.GetClient("", "")
	if client == nil || err != nil {
		glog.Errorf("client init failed, %s", err.Error())
		return
	}

	UnpublishAttachmentList = NewList()
	go UnpublishRoutine()
}

// GetDefaultProfile implementation
func GetDefaultProfile() (*model.ProfileSpec, error) {
	profiles, err := client.ListProfiles()
	if err != nil {
		glog.Error("Get default profile failed: ", err)
		return nil, err
	}

	for _, profile := range profiles {
		if profile.Name == "default" {
			return profile, nil
		}
	}

	return nil, status.Error(codes.FailedPrecondition, "No default profile")
}

// FindVolume implementation
func FindVolume(req *model.VolumeSpec) (bool, bool, *model.VolumeSpec, error) {
	isExist := false
	volumes, err := client.ListVolumes()

	if err != nil {
		glog.Error("List volumes failed: ", err)

		return false, false, nil, err
	}

	for _, volume := range volumes {
		if volume.Name == req.Name {
			isExist = true

			if (volume.Size == req.Size) && (volume.ProfileId == req.ProfileId) &&
				(volume.AvailabilityZone == req.AvailabilityZone) &&
				(volume.SnapshotId == req.SnapshotId) {
				glog.V(5).Infof("Volume already exists and is compatible")

				return true, true, volume, nil
			}
		}
	}

	return isExist, false, nil, nil
}

// CreateVolume implementation
func (p *Plugin) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	glog.V(5).Info("start to CreateVolume")
	defer glog.V(5).Info("end to CreateVolume")

	if req.Name == "" {
		msg := "CreateVolume Name must be provided"
		glog.Errorf("CreateVolume Name must be provided")
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		msg := "CreateVolume Volume capabilities must be provided"
		glog.Errorf(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	// build volume body
	var fstype string
	volumebody := &model.VolumeSpec{}
	volumebody.Name = req.Name
	var secondaryAZ = util.OpensdsDefaultSecondaryAZ
	var enableReplication = false

	for k, v := range req.GetParameters() {
		switch strings.ToLower(k) {
		case KVolumeFstype:
			fstype = v

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
		case KMultiAttach:
			if strings.ToLower(v) == "true" {
				volumebody.MultiAttach = true
			}
		}
	}

	if !util.IsSupportFstype(fstype) {
		msg := (fmt.Sprintf("Volume create fstype[%s] not support.", fstype))
		glog.Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	allocationUnitBytes := util.GiB
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

	contentSource := req.GetVolumeContentSource()
	if nil != contentSource {
		snapshot := contentSource.GetSnapshot()
		if snapshot != nil {
			volumebody.SnapshotId = snapshot.GetSnapshotId()
		}
	}

	if "" == volumebody.ProfileId {
		defaultRrf, err := GetDefaultProfile()
		if err != nil {
			return nil, err
		}

		volumebody.ProfileId = defaultRrf.Id
	}

	if "" == volumebody.AvailabilityZone {
		volumebody.AvailabilityZone = "default"
	}

	glog.V(5).Infof("CreateVolume volumebody: %v", volumebody)

	isExist, isCompatible, findVolume, err := FindVolume(volumebody)
	if err != nil {
		return nil, err
	}

	var v *model.VolumeSpec

	if isExist {
		if isCompatible {
			v = findVolume
		} else {
			return nil, status.Error(codes.AlreadyExists,
				"Volume already exists but is incompatible")
		}
	} else {
		createVolume, err := client.CreateVolume(volumebody)
		if err != nil {
			glog.Error("failed to CreateVolume", err)
			return nil, err
		} else {
			v = createVolume
		}
	}

	glog.V(5).Infof("waiting until volume is created.")
	volStable, err := p.waitForVolStatusStable(v.Id)
	if err != nil {
		msg := fmt.Sprintf("Failed to CreateVolume:errMsg: %v", err)
		glog.Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	if volStable.Status != "available" {
		msg := fmt.Sprintf("Failed to CreateVolume: volume %s status %s is invalid.", volStable.Id, volStable.Status)
		glog.Errorf(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	glog.V(5).Infof("opensds volume = %v", v)

	// return volume info
	volumeinfo := &csi.Volume{
		CapacityBytes: v.Size * allocationUnitBytes,
		VolumeId:      v.Id,
		VolumeContext: map[string]string{
			KVolumeName:      v.Name,
			KVolumeStatus:    v.Status,
			KVolumeAZ:        v.AvailabilityZone,
			KVolumePoolId:    v.PoolId,
			KVolumeProfileId: v.ProfileId,
			KVolumeLvPath:    v.Metadata["lvPath"],
			KVolumeFstype:    fstype,
		},
	}

	glog.V(5).Infof("resp volumeinfo = %v", volumeinfo)
	if enableReplication && !isExist {
		volumebody.AvailabilityZone = secondaryAZ
		volumebody.Name = SecondaryPrefix + req.Name
		sVol, err := client.CreateVolume(volumebody)
		if err != nil {
			glog.Errorf("failed to create secondar volume: %v", err)
			return nil, err
		}

		sVolStable, err := p.waitForVolStatusStable(sVol.Id)
		if err != nil {
			msg := fmt.Sprintf("Failed to CreateVolume:errMsg: %v", err)
			glog.Errorf(msg)
			return nil, status.Error(codes.Internal, msg)
		}
		if sVolStable.Status != "available" {
			msg := fmt.Sprintf("Failed to CreateVolume: volume %s status %s is invalid.", sVolStable.Id, sVolStable.Status)
			glog.Errorf(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		replicaBody := &model.ReplicationSpec{
			Name:              req.Name,
			PrimaryVolumeId:   v.Id,
			SecondaryVolumeId: sVol.Id,
			ReplicationMode:   model.ReplicationModeSync,
			ReplicationPeriod: 0,
		}
		replicaResp, err := client.CreateReplication(replicaBody)
		if err != nil {
			glog.Errorf("Create replication failed: %v", err)
			return nil, err
		}
		volumeinfo.VolumeContext[KVolumeReplicationId] = replicaResp.Id
	}

	return &csi.CreateVolumeResponse{
		Volume: volumeinfo,
	}, nil
}

func getReplicationByVolume(volId string) *model.ReplicationSpec {
	replications, _ := client.ListReplications()
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

	if req.VolumeId == "" {
		msg := "DeleteVolume Volume ID must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	volId := req.VolumeId

	vol, err := client.GetVolume(volId)
	if err != nil {
		glog.Error("Get volume failed, ", err)
		return nil, err
	}

	if vol == nil {
		return nil, fmt.Errorf("The volume %s is already deleted.", volId)
	}

	r := getReplicationByVolume(volId)
	if r != nil {
		if err := client.DeleteReplication(r.Id, nil); err != nil {
			glog.Error("Delete replication failed, ", err)
			return nil, err
		}
		if err := client.DeleteVolume(r.PrimaryVolumeId, &model.VolumeSpec{}); err != nil {
			glog.Error("Delete primary volume failed, ", err)
			return nil, err
		}
		if err := client.DeleteVolume(r.SecondaryVolumeId, &model.VolumeSpec{}); err != nil {
			glog.Error("Delete secondary volume failed, ", err)
			return nil, err
		}
	} else {
		if err := client.DeleteVolume(volId, &model.VolumeSpec{}); err != nil {
			glog.Error("Delete volume failed, ", err)
			return nil, err
		}
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// isStringMapEqual implementation
func isStringMapEqual(metadataA, metadataB map[string]string) bool {
	glog.V(5).Infof("start to isStringMapEqual, metadataA = %v, metadataB = %v!",
		metadataA, metadataB)
	if len(metadataA) != len(metadataB) {
		glog.V(5).Infof("len(metadataA)(%v) != len(metadataB)(%v) ",
			len(metadataA), len(metadataB))
		return false
	}

	for key, valueA := range metadataA {
		valueB, ok := metadataB[key]
		if !ok || (valueA != valueB) {
			glog.V(5).Infof("ok = %v, key = %v, valueA = %v, valueB = %v!",
				ok, key, valueA, valueB)
			return false
		}
	}

	return true
}

// isVolumePublished Check if the volume is published and compatible
func isVolumePublished(canAtMultiNode bool, attachReq *model.VolumeAttachmentSpec,
	metadata map[string]string) (*model.VolumeAttachmentSpec, error) {
	glog.V(5).Infof("start to isVolumePublished, canAtMultiNode = %v, attachReq = %v",
		canAtMultiNode, attachReq)

	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		glog.V(5).Info("ListVolumeAttachments failed: " + err.Error())
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == attachReq.VolumeId {
			if attachSpec.Host != attachReq.Host {
				if !canAtMultiNode {
					msg := fmt.Sprintf("the volume %s has been published to another node and does not have MULTI_NODE volume capability",
						attachReq.VolumeId)
					return nil, status.Error(codes.FailedPrecondition, msg)
				}
			} else {
				// Opensds does not have volume_capability and readonly parameters,
				// but needs to check other parameters to determine compatibility?
				if attachSpec.Platform == attachReq.Platform &&
					attachSpec.OsType == attachReq.OsType &&
					attachSpec.Initiator == attachReq.Initiator &&
					isStringMapEqual(attachSpec.Metadata, metadata) &&
					attachSpec.AccessProtocol == attachReq.AccessProtocol {
					glog.V(5).Info("Volume published and is compatible")

					return attachSpec, nil
				}

				glog.Error("Volume published but is incompatible, incompatible attachement Id = " + attachSpec.Id)
				return nil, status.Error(codes.AlreadyExists, "Volume published but is incompatible")
			}
		}
	}

	glog.V(5).Info("Need to create a new attachment")
	return nil, nil
}

// ControllerPublishVolume implementation
func (p *Plugin) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	glog.V(5).Info("start to ControllerPublishVolume")
	defer glog.V(5).Info("end to ControllerPublishVolume")

	if req.VolumeId == "" {
		msg := "ControllerPublishVolume Volume ID must be provided"
		glog.Info(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if req.NodeId == "" {
		msg := "ControllerPublishVolume Node ID must be provided"
		glog.Info(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	fstype, ok := req.VolumeContext[KVolumeFstype]
	if !ok {
		msg := "ControllerPublishVolume fstype must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	//check volume is exist
	volSpec, errVol := client.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", req.VolumeId)
		return nil, status.Error(codes.NotFound, msg)
	}

	pool, err := client.GetPool(volSpec.PoolId)
	if err != nil || pool == nil {
		msg := fmt.Sprintf("the pool %s is not sxist", volSpec.PoolId)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	var protocol = strings.ToLower(pool.Extras.IOConnectivity.AccessProtocol)

	var initator string
	var nodeInfo = req.NodeId

	switch protocol {
	case connector.FcDriver:
		wwpns, err := extractFCInitiatorFromNodeInfo(nodeInfo)
		if err != nil {
			glog.Error(err.Error())
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		initator = strings.Join(wwpns, ",")
		break
	case connector.IscsiDriver:
		iqn, err := extractISCSIInitiatorFromNodeInfo(nodeInfo)
		if err != nil {
			glog.Error(err.Error())
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		initator = iqn
		break
	case connector.RbdDriver:
		break
	default:
		msg := fmt.Sprintf("protocol:[%s] not support.", protocol)
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	attachReq := &model.VolumeAttachmentSpec{
		VolumeId: req.VolumeId,
		HostInfo: model.HostInfo{
			Host:      strings.Split(nodeInfo, ",")[0],
			Platform:  runtime.GOARCH,
			OsType:    runtime.GOOS,
			Initiator: initator,
		},
		Metadata:       req.VolumeContext,
		AccessProtocol: protocol,
	}

	mode := req.VolumeCapability.AccessMode.Mode
	canAtMultiNode := false

	if csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER == mode {
		canAtMultiNode = true
	}

	expectedMetadata := utils.MergeStringMaps(attachReq.Metadata, volSpec.Metadata)
	existAttachment, err := isVolumePublished(canAtMultiNode, attachReq, expectedMetadata)
	if err != nil {
		return nil, err
	}

	var attachSpec *model.VolumeAttachmentSpec

	if nil == existAttachment {
		newAttachment, errAttach := client.CreateVolumeAttachment(attachReq)
		if errAttach != nil {
			msg := fmt.Sprintf("the volume %s failed to publish to node %s.", req.VolumeId, req.NodeId)
			glog.Errorf("failed to ControllerPublishVolume: %v", attachReq)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		attachSpec = newAttachment
	} else {
		attachSpec = existAttachment
	}

	resp := &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			KPublishHostIp:       attachSpec.Ip,
			KPublishHostName:     attachSpec.Host,
			KPublishAttachId:     attachSpec.Id,
			KPublishAttachStatus: attachSpec.Status,
			KVolumeFstype:        fstype,
		},
	}

	if replicationId, ok := req.VolumeContext[KVolumeReplicationId]; ok {
		r, err := client.GetReplication(replicationId)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, "Get replication failed")
		}

		attachReq.VolumeId = r.SecondaryVolumeId
		existAttachment, err := isVolumePublished(canAtMultiNode, attachReq, expectedMetadata)
		if err != nil {
			return nil, err
		}

		if nil == existAttachment {
			newAttachment, errAttach := client.CreateVolumeAttachment(attachReq)
			if errAttach != nil {
				msg := fmt.Sprintf("the volume %s failed to publish to node %s.", req.VolumeId, req.NodeId)
				glog.Errorf("failed to ControllerPublishVolume: %v", attachReq)
				return nil, status.Error(codes.FailedPrecondition, msg)
			}

			attachSpec = newAttachment
		} else {
			attachSpec = existAttachment
		}

		resp.PublishContext[KPublishSecondaryAttachId] = attachSpec.Id
	}
	return resp, nil
}

// ControllerUnpublishVolume implementation
func (p *Plugin) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	glog.V(5).Infof("start to ControllerUnpublishVolume, req VolumeId = %v, NodeId = %v, ControllerUnpublishSecrets =%v",
		req.VolumeId, req.NodeId, req.Secrets)
	defer glog.V(5).Info("end to ControllerUnpublishVolume")

	if req.VolumeId == "" {
		msg := "ControllerPublishVolume Volume ID must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if req.NodeId == "" {
		msg := "ControllerUnpublishVolume Node ID must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	//check volume is exist
	volSpec, errVol := client.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", req.VolumeId)
		return nil, status.Error(codes.NotFound, msg)
	}

	if volSpec.Status == model.VolumeAvailable {
		msg := fmt.Sprintf("volume %s has already been unpublished.", volSpec.Id)
		glog.Error(msg)
		return &csi.ControllerUnpublishVolumeResponse{}, nil
	}

	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to unpublish volume.")
	}

	hostName := strings.Split(req.NodeId, ",")[0]
	var acts []*model.VolumeAttachmentSpec

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == req.VolumeId && (req.NodeId == "" || attachSpec.Host == hostName) {
			acts = append(acts, attachSpec)
		}
	}

	if r := getReplicationByVolume(req.VolumeId); r != nil {
		for _, attachSpec := range attachments {
			if attachSpec.VolumeId == r.SecondaryVolumeId && (req.NodeId == "" || attachSpec.Host == hostName) {
				acts = append(acts, attachSpec)
			}
		}
	}

	for _, act := range acts {
		if ok := UnpublishAttachmentList.isExist(act.Id); !ok {
			glog.Infof("Add attachment id %s into unpublish attachment list.", act.Id)
			UnpublishAttachmentList.Add(act)
			UnpublishAttachmentList.PrintList()
		}
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities implementation
func (p *Plugin) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes implementation
func (p *Plugin) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	glog.V(5).Info("start to ListVolumes")
	defer glog.V(5).Info("end to ListVolumes")

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	// only support list all the volumes at present
	volumes, err := client.ListVolumes()
	if err != nil {
		return nil, err
	}

	ens := []*csi.ListVolumesResponse_Entry{}
	for _, v := range volumes {
		if v != nil {

			volumeinfo := &csi.Volume{
				CapacityBytes: v.Size,
				VolumeId:      v.Id,
				VolumeContext: map[string]string{
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

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	pools, err := client.ListPools()
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

// FindSnapshot implementation
func FindSnapshot(req *model.VolumeSnapshotSpec) (bool, bool, *model.VolumeSnapshotSpec, error) {
	isExist := false
	snapshots, err := client.ListVolumeSnapshots()

	if err != nil {
		glog.Error("List volume snapshots failed: ", err)

		return false, false, nil, err
	}

	for _, snapshot := range snapshots {
		if snapshot.Name == req.Name {
			isExist = true

			if (snapshot.VolumeId == req.VolumeId) && (snapshot.ProfileId == req.ProfileId) {
				glog.V(5).Infof("snapshot already exists and is compatible")

				return true, true, snapshot, nil
			}
		}
	}

	return isExist, false, nil, nil
}

// CreateSnapshot implementation
func (p *Plugin) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {

	defer glog.V(5).Info("end to CreateSnapshot")
	glog.V(5).Infof("start to CreateSnapshot, Name: %v, SourceVolumeId: %v, CreateSnapshotSecrets: %v, parameters: %v!",
		req.Name, req.SourceVolumeId, req.Secrets, req.Parameters)

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

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

	for k, v := range req.GetParameters() {
		switch strings.ToLower(k) {
		// TODO: support profile name
		case KParamProfile:
			snapReq.ProfileId = v
		}
	}

	glog.Infof("opensds CreateVolumeSnapshot request body: %v", snapReq)
	var snapshot *model.VolumeSnapshotSpec
	isExist, isCompatible, findSnapshot, err := FindSnapshot(snapReq)

	if err != nil {
		return nil, err
	}

	if isExist {
		if isCompatible {
			snapshot = findSnapshot
		} else {
			return nil, status.Error(codes.AlreadyExists,
				"Snapshot already exists but is incompatible")
		}
	} else {
		createSnapshot, err := client.CreateVolumeSnapshot(snapReq)
		if err != nil {
			glog.Error("failed to CreateVolumeSnapshot", err)
			return nil, err
		}

		snapshot = createSnapshot
	}

	glog.V(5).Infof("opensds snapshot = %v", snapshot)
	creationTime, err := p.convertStringToPtypesTimestamp(snapshot.CreatedAt)
	if nil != err {
		return nil, err
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      snapshot.Size * util.GiB,
			SnapshotId:     snapshot.Id,
			SourceVolumeId: snapshot.VolumeId,
			CreationTime:   creationTime,
			ReadyToUse:     true,
		},
	}, nil
}

func (p *Plugin) convertStringToPtypesTimestamp(timeStr string) (*timestamp.Timestamp, error) {
	timeAt, err := time.Parse(constants.TimeFormat, timeStr)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ptypesTime, err := ptypes.TimestampProto(timeAt)
	if err != nil {
		return nil, err
	}
	return ptypesTime, nil
}

// DeleteSnapshot implementation
func (p *Plugin) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {

	defer glog.V(5).Info("end to DeleteSnapshot")
	glog.V(5).Infof("start to DeleteSnapshot, SnapshotId: %v, DeleteSnapshotSecrets: %v!",
		req.SnapshotId, req.Secrets)

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	if 0 == len(req.SnapshotId) {
		return nil, status.Error(codes.InvalidArgument, "Snapshot ID cannot be empty")
	}

	err := client.DeleteVolumeSnapshot(req.SnapshotId, nil)

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

	if client == nil {
		return nil, status.Error(codes.InvalidArgument, "client is nil")
	}

	var opts map[string]string
	allSnapshots, err := client.ListVolumeSnapshots(opts)
	if nil != err {
		return nil, err
	}

	snapshotId := req.GetSnapshotId()
	snapshotIDLen := len(snapshotId)
	sourceVolumeId := req.GetSourceVolumeId()
	sourceVolumeIdLen := len(sourceVolumeId)
	var snapshotsFilterByVolumeId []*model.VolumeSnapshotSpec
	var snapshotsFilterById []*model.VolumeSnapshotSpec
	var filterResult []*model.VolumeSnapshotSpec

	for _, snapshot := range allSnapshots {
		if snapshot.VolumeId == sourceVolumeId {
			snapshotsFilterByVolumeId = append(snapshotsFilterByVolumeId, snapshot)
		}

		if snapshot.Id == snapshotId {
			snapshotsFilterById = append(snapshotsFilterById, snapshot)
		}
	}

	switch {
	case (0 == snapshotIDLen) && (0 == sourceVolumeIdLen):
		if len(allSnapshots) <= 0 {
			glog.V(5).Info("len(allSnapshots) <= 0")
			return &csi.ListSnapshotsResponse{}, nil
		}

		filterResult = allSnapshots
		break
	case (0 == snapshotIDLen) && (0 != sourceVolumeIdLen):
		if len(snapshotsFilterByVolumeId) <= 0 {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("no snapshot with source volume id %s", sourceVolumeId))
		}

		filterResult = snapshotsFilterByVolumeId
		break
	case (0 != snapshotIDLen) && (0 == sourceVolumeIdLen):
		if len(snapshotsFilterById) <= 0 {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("no snapshot with id %s", snapshotId))
		}

		filterResult = snapshotsFilterById
		break
	case (0 != snapshotIDLen) && (0 != sourceVolumeIdLen):
		for _, snapshot := range snapshotsFilterById {
			if snapshot.VolumeId == sourceVolumeId {
				filterResult = append(filterResult, snapshot)
			}
		}

		if len(filterResult) <= 0 {
			return nil, status.Error(codes.NotFound,
				fmt.Sprintf("no snapshot with id %v and source volume id %v", snapshotId, sourceVolumeId))
		}

		break
	}

	glog.V(5).Infof("filterResult=%v.", filterResult)
	var sortedKeys []string
	snapshotsMap := make(map[string]*model.VolumeSnapshotSpec)

	for _, snapshot := range filterResult {
		sortedKeys = append(sortedKeys, snapshot.Id)
		snapshotsMap[snapshot.Id] = snapshot
	}
	sort.Strings(sortedKeys)

	var sortResult []*model.VolumeSnapshotSpec
	for _, key := range sortedKeys {
		sortResult = append(sortResult, snapshotsMap[key])
	}

	var (
		ulenSnapshots = int32(len(sortResult))
		maxEntries    = req.MaxEntries
		startingToken int32
	)

	if v := req.StartingToken; v != "" {
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, status.Error(codes.Aborted, "parsing the startingToken failed")
		}
		startingToken = int32(i)
	}

	if startingToken >= ulenSnapshots {
		return nil, status.Errorf(
			codes.Aborted,
			"startingToken=%d >= len(snapshots)=%d",
			startingToken, ulenSnapshots)
	}

	// If maxEntries is 0 or greater than the number of remaining entries then
	// set maxEntries to the number of remaining entries.
	var sliceResult []*model.VolumeSnapshotSpec
	var nextToken string
	nextTokenIndex := startingToken + maxEntries

	if maxEntries == 0 || nextTokenIndex >= ulenSnapshots {
		sliceResult = sortResult[startingToken:]
	} else {
		sliceResult = sortResult[startingToken:nextTokenIndex]
		nextToken = fmt.Sprintf("%d", nextTokenIndex)
	}

	glog.V(5).Infof("sliceResult=%v, nextToken=%v.", sliceResult, nextToken)
	if len(sliceResult) <= 0 {
		return &csi.ListSnapshotsResponse{NextToken: nextToken}, nil
	}

	entries := []*csi.ListSnapshotsResponse_Entry{}
	for _, snapshot := range sliceResult {
		creationTime, err := p.convertStringToPtypesTimestamp(snapshot.CreatedAt)
		if nil != err {
			return nil, err
		}
		entries = append(entries, &csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SizeBytes:      snapshot.Size * util.GiB,
				SnapshotId:     snapshot.Id,
				SourceVolumeId: snapshot.VolumeId,
				CreationTime:   creationTime,
				ReadyToUse:     true,
			},
		})
	}

	glog.V(5).Infof("entries=%v.", entries)
	return &csi.ListSnapshotsResponse{
		Entries:   entries,
		NextToken: nextToken,
	}, nil
}

func (p *Plugin) waitForVolStatusStable(volumeID string) (*model.VolumeSpec, error) {

	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(5 * time.Minute)

	defer ticker.Stop()
	validVolumeStatus := []string{"error", "error_deleting", "error_restoring", "error_extending", "available", "in-use"}

	for {
		select {
		case <-ticker.C:
			vol, err := client.GetVolume(volumeID)
			if err != nil {
				return nil, fmt.Errorf("Get volume %s failed, errInfo: %v", volumeID, err)
			}

			if vol != nil && util.Contained(vol.Status, validVolumeStatus) {
				return vol, nil
			}

		case <-timeout:
			return nil, fmt.Errorf("timeout occured waiting for checking status of the volume %s", volumeID)
		}
	}
}

// AttachmentObj implementation
type AttachmentObj struct {
	l *list.List
	m sync.Mutex
	r sync.RWMutex
}

// NewList implementation
func NewList() *AttachmentObj {
	return &AttachmentObj{l: list.New()}
}

// UnpublishAttachmentList implementation
var UnpublishAttachmentList *AttachmentObj

// Add implementation
func (q *AttachmentObj) Add(v interface{}) {
	if v == nil {
		return
	}
	q.m.Lock()
	defer q.m.Unlock()
	q.l.PushBack(v)
}

// GetHead implementation
func (q *AttachmentObj) GetHead() *list.Element {
	q.r.RLock()
	defer q.r.RUnlock()
	return q.l.Front()
}

// isExist implementation
func (q *AttachmentObj) isExist(v interface{}) bool {
	if q.GetLen() == 0 {
		return false
	}
	for e := q.GetHead(); e != nil; e = e.Next() {
		if e.Value == v {
			return true
		}
	}
	return false
}

// Delete implementation
func (q *AttachmentObj) Delete(e *list.Element) {
	if e == nil {
		return
	}
	q.m.Lock()
	defer q.m.Unlock()
	q.l.Remove(e)
}

// GetLen implementation
func (q *AttachmentObj) GetLen() int {
	q.r.RLock()
	defer q.r.RUnlock()
	return q.l.Len()
}

// PrintList implementation
func (q *AttachmentObj) PrintList() {
	var attachmentIDList string
	for e := q.GetHead(); e != nil; e = e.Next() {
		attachmentIDList = attachmentIDList + e.Value.(*model.VolumeAttachmentSpec).Id + ","
	}
	glog.Infof("The list of attachments in the context is %s", attachmentIDList)
}

// UnpublishRoutine implementation
func UnpublishRoutine() {
	for {
		listLen := UnpublishAttachmentList.GetLen()
		if listLen > 0 {
			var next *list.Element
			for e := UnpublishAttachmentList.GetHead(); e != nil; e = next {
				next = e.Next()
				act := e.Value.(*model.VolumeAttachmentSpec)

				err := client.DeleteVolumeAttachment(act.Id, act)
				if err != nil {
					glog.Errorf("the volume %s failed to unpublish from node %s, error: %v.", act.VolumeId, act.Host, err)
				} else {
					waitVolumeAttachmentDeleted(act, e)
				}
				time.Sleep(10 * time.Second)
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func waitVolumeAttachmentDeleted(act *model.VolumeAttachmentSpec, e *list.Element) {
	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ticker.C:
			attachment, _ := client.GetVolumeAttachment(act.Id)
			if attachment != nil {
				glog.Errorf("Waiting for the volume: %s successfully to unpublish to node: %s", act.VolumeId, act.Host)
			} else {
				glog.V(5).Infof("The volume: %s successfully to unpublish to node: %s", act.VolumeId, act.Host)
				UnpublishAttachmentList.Delete(e)
				return
			}

		case <-timeout:
			glog.Errorf("timeout occured waiting for checking deletion of the volume attachment %s", act.Id)
			return
		}
	}
}

func extractISCSIInitiatorFromNodeInfo(nodeInfo string) (string, error) {
	for _, v := range strings.Split(nodeInfo, ",") {
		if strings.Contains(v, "iqn") {
			glog.Info("ISCSI initiator is ", v)
			return v, nil
		}
	}

	return "", errors.New("No ISCSI initiators found")
}

func extractFCInitiatorFromNodeInfo(nodeInfo string) ([]string, error) {
	var wwpns []string
	for _, v := range strings.Split(nodeInfo, ",") {
		if strings.Contains(v, "node_name") {
			wwpns = append(wwpns, strings.Split(v, ":")[1])
		}
	}

	if len(wwpns) == 0 {
		return nil, errors.New("No FC initiators found.")
	}

	glog.Info("FC initiators are ", wwpns)

	return wwpns, nil
}
