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
	"github.com/opensds/nbp/csi/util"
	"github.com/opensds/opensds/contrib/connector"
	"github.com/opensds/opensds/pkg/model"
	"github.com/opensds/opensds/pkg/utils/constants"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

func init() {
	UnpublishAttachmentList = NewList()
	c := &Plugin{}
	go c.UnpublishRoutine()
}

// GetDefaultProfile implementation
func (c *Plugin) GetDefaultProfile() (*model.ProfileSpec, error) {
	profiles, err := c.Cli.ListProfiles()
	if err != nil {
		glog.Errorf("get default profile failed: %v", err)
		return nil, err
	}

	for _, profile := range profiles {
		if profile.Name == "default" {
			return profile, nil
		}
	}

	return nil, status.Error(codes.FailedPrecondition, "no default profile")
}

// FindVolume implementation
func (c *Plugin) FindVolume(req *model.VolumeSpec) (*model.VolumeSpec, error) {
	volumes, err := c.Cli.ListVolumes()
	if err != nil {
		msg := fmt.Sprintf("list volumes failed: %v", err)
		glog.Error(msg)
		return nil, errors.New(msg)
	}

	for _, volume := range volumes {
		if volume.Name == req.Name {
			return volume, nil
		}
	}

	return nil, nil
}

// CreateVolume implementation
func (c *Plugin) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	glog.V(5).Info("start to create volume")
	defer glog.V(5).Info("end to create volume")

	if req.Name == "" {
		msg := "volume name must be provided when creating volume"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		msg := "volume capabilities must be provided when creating volume"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	// build volume body
	var fstype string
	volumebody := &model.VolumeSpec{}
	volumebody.Name = req.Name
	var secondaryAZ = util.OpensdsDefaultSecondaryAZ
	var enableReplication = false
	var attachMode = "rw"
	glog.V(5).Infof("create volume parameters %+v", req.GetParameters())
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
		case KPublishAttachMode:
			if strings.ToLower(v) == "ro" {
				attachMode = "ro"
			}
		}
	}

	if !util.IsSupportFstype(fstype) {
		msg := fmt.Sprintf("volume creation fstype: %s is not support.", fstype)
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
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
		defaultRrf, err := c.GetDefaultProfile()
		if err != nil {
			return nil, err
		}

		volumebody.ProfileId = defaultRrf.Id
	}

	if "" == volumebody.AvailabilityZone {
		volumebody.AvailabilityZone = "default"
	}

	glog.V(5).Infof("volume body: %+v", volumebody)

	volExist, err := c.FindVolume(volumebody)
	if err != nil {
		return nil, err
	}

	var v *model.VolumeSpec

	if volExist == nil {
		createVolume, err := c.Cli.CreateVolume(volumebody)
		if err != nil {
			msg := fmt.Sprintf("failed to create volume: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		v = createVolume

	} else {
		v = volExist
	}

	glog.V(5).Info("waiting until volume is created")
	volStable, err := c.waitForVolStatusStable(v.Id)
	if err != nil {
		msg := fmt.Sprintf("failed to create volume: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	if volStable.Status != "available" {
		msg := fmt.Sprintf("failed to create volume: volume %s status %s is invalid", volStable.Id, volStable.Status)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	glog.V(5).Infof("opensds volume = %+v", v)

	// return volume info
	volumeinfo := &csi.Volume{
		CapacityBytes: v.Size * allocationUnitBytes,
		VolumeId:      v.Id,
		VolumeContext: map[string]string{
			KVolumeName:        v.Name,
			KVolumeStatus:      v.Status,
			KVolumeAZ:          v.AvailabilityZone,
			KVolumePoolId:      v.PoolId,
			KVolumeProfileId:   v.ProfileId,
			KVolumeLvPath:      v.Metadata["lvPath"],
			KVolumeFstype:      fstype,
			KPublishAttachMode: attachMode,
		},
	}

	glog.V(5).Infof("response volume info = %+v", volumeinfo)
	if enableReplication && volExist == nil {
		volumebody.AvailabilityZone = secondaryAZ
		volumebody.Name = SecondaryPrefix + req.Name
		sVol, err := c.Cli.CreateVolume(volumebody)
		if err != nil {
			msg := fmt.Sprintf("failed to create second volume: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		sVolStable, err := c.waitForVolStatusStable(sVol.Id)
		if err != nil {
			msg := fmt.Sprintf("failed to create volume: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
		if sVolStable.Status != "available" {
			msg := fmt.Sprintf("failed to create volume: volume %s status %s is invalid.", sVolStable.Id, sVolStable.Status)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		replicaBody := &model.ReplicationSpec{
			Name:              req.Name,
			PrimaryVolumeId:   v.Id,
			SecondaryVolumeId: sVol.Id,
			ReplicationMode:   model.ReplicationModeSync,
			ReplicationPeriod: 0,
		}
		replicaResp, err := c.Cli.CreateReplication(replicaBody)
		if err != nil {
			msg := fmt.Sprintf("create replication failed: %v", err)
			glog.Errorf(msg)
			return nil, status.Error(codes.Internal, msg)
		}
		volumeinfo.VolumeContext[KVolumeReplicationId] = replicaResp.Id
	}

	return &csi.CreateVolumeResponse{
		Volume: volumeinfo,
	}, nil
}

func (c *Plugin) getReplicationByVolume(volId string) *model.ReplicationSpec {
	replications, _ := c.Cli.ListReplications()
	for _, r := range replications {
		if volId == r.PrimaryVolumeId || volId == r.SecondaryVolumeId {
			return r
		}
	}
	return nil
}

// DeleteVolume implementation
func (c *Plugin) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	glog.V(5).Info("start to delete volume")
	defer glog.V(5).Info("end to delete volume")

	if req.VolumeId == "" {
		msg := "volume ID must be provided when deleting volume"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	volId := req.VolumeId

	vol, err := c.Cli.GetVolume(volId)
	if err != nil {
		msg := fmt.Sprintf("get volume failed: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if vol == nil {
		return &csi.DeleteVolumeResponse{}, nil
	}

	r := c.getReplicationByVolume(volId)
	if r != nil {
		if err := c.Cli.DeleteReplication(r.Id, nil); err != nil {
			msg := fmt.Sprintf("delete replication failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
		if err := c.Cli.DeleteVolume(r.PrimaryVolumeId, &model.VolumeSpec{}); err != nil {
			msg := fmt.Sprintf("delete primary volume failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
		if err := c.Cli.DeleteVolume(r.SecondaryVolumeId, &model.VolumeSpec{}); err != nil {
			msg := fmt.Sprintf("delete secondary volume failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
	} else {
		if err := c.Cli.DeleteVolume(volId, &model.VolumeSpec{}); err != nil {
			msg := fmt.Sprintf("delete volume failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// isStringMapEqual implementation
func isStringMapEqual(metadataA, metadataB map[string]string) bool {
	glog.V(5).Infof("start to determine whether the two map structures are equal, metadataA = %v, metadataB = %v!",
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
func (c *Plugin) isVolumeCanBePublished(canAtMultiNode bool, attachReq *model.VolumeAttachmentSpec,
	volMultiAttach bool) error {

	glog.V(5).Infof("start to judge whether volume can be published, canAtMultiNode = %v, attachReq = %v",
		canAtMultiNode, attachReq)

	attachments, err := c.Cli.ListVolumeAttachments()
	if err != nil {
		msg := fmt.Sprintf("list volume attachments failed: %v", err)
		glog.Error(msg)
		return status.Error(codes.FailedPrecondition, msg)
	}

	msg := fmt.Sprintf("the volume %s can be published", attachReq.VolumeId)

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == attachReq.VolumeId {
			if attachSpec.Host == attachReq.Host {
				msg := fmt.Sprintf("the volume %s is publishing to the current node %s and no need to publish again", attachReq.VolumeId, attachReq.Host)
				glog.Infof(msg)
				return nil
			}
			if !canAtMultiNode {
				msg := fmt.Sprintf("the volume %s has been published to the node %s and kubernetes does not have MULTI_NODE volume capability", attachReq.VolumeId, attachSpec.Host)
				glog.Error(msg)
				return status.Error(codes.FailedPrecondition, msg)
			}
			if !volMultiAttach {
				msg := fmt.Sprintf("the volume %s has been published to the node %s, but the volume does not enable multiattach", attachReq.VolumeId, attachSpec.Host)
				glog.Error(msg)
				return status.Error(codes.FailedPrecondition, msg)
			}

			glog.Info(msg)
			return nil
		}
	}

	glog.Info(msg)
	return nil
}

// ControllerPublishVolume implementation
func (c *Plugin) ControllerPublishVolume(ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	glog.V(5).Info("start to controller publish volume")
	defer glog.V(5).Info("end to controller publish volume")

	if req.VolumeId == "" {
		msg := "volume ID must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if req.NodeId == "" {
		msg := "node ID must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	fstype, ok := req.VolumeContext[KVolumeFstype]
	if !ok {
		msg := "fstype must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	attachMode, ok := req.VolumeContext[KPublishAttachMode]
	if !ok {
		glog.Info("attach mode will use default value: rw")
		attachMode = "rw"
	}

	//check volume is exist
	volSpec, err := c.Cli.GetVolume(req.VolumeId)
	if err != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist: %v",
			req.VolumeId, err)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	pool, err := c.Cli.GetPool(volSpec.PoolId)
	if err != nil || pool == nil {
		msg := fmt.Sprintf("the pool %s is not exist: %v",
			volSpec.PoolId, err)
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
			msg := fmt.Sprintf("extract FC initiator from node info failed: %v",
				err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		initator = strings.Join(wwpns, ",")
		break
	case connector.IscsiDriver:
		iqn, err := extractISCSIInitiatorFromNodeInfo(nodeInfo)
		if err != nil {
			msg := fmt.Sprintf("extract ISCSI initiator from node info failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		initator = iqn
		break
	case connector.RbdDriver:
		break
	default:
		msg := fmt.Sprintf("protocol:%s not support", protocol)
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
		AttachMode:     attachMode,
	}

	mode := req.VolumeCapability.AccessMode.Mode
	canAtMultiNode := false

	if csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER == mode {
		canAtMultiNode = true
	}

	err = c.isVolumeCanBePublished(canAtMultiNode, attachReq, volSpec.MultiAttach)
	if err != nil {
		return nil, err
	}

	newAttachment, errAttach := c.Cli.CreateVolumeAttachment(attachReq)
	if errAttach != nil {
		msg := fmt.Sprintf("the volume %s is failed to publish to node %s, error info: %v",
			req.VolumeId, req.NodeId, errAttach)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	resp := &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			KPublishHostIp:       newAttachment.Ip,
			KPublishHostName:     newAttachment.Host,
			KPublishAttachId:     newAttachment.Id,
			KPublishAttachStatus: newAttachment.Status,
			KVolumeFstype:        fstype,
			KPublishAttachMode:   attachMode,
		},
	}

	if replicationId, ok := req.VolumeContext[KVolumeReplicationId]; ok {
		r, err := c.Cli.GetReplication(replicationId)
		if err != nil {
			msg := fmt.Sprintf("failed to get replication: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		attachReq.VolumeId = r.SecondaryVolumeId

		secondaryVolume, err := c.Cli.GetVolume(attachReq.VolumeId)
		if err != nil {
			msg := fmt.Sprintf("failed to get secondary volume: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}

		err = c.isVolumeCanBePublished(canAtMultiNode, attachReq, secondaryVolume.MultiAttach)
		if err != nil {
			return nil, err
		}

		newAttachment, errAttach := c.Cli.CreateVolumeAttachment(attachReq)
		if errAttach != nil {
			msg := fmt.Sprintf("the volume %s failed to publish to node %s, error info %v",
				req.VolumeId, req.NodeId, errAttach)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
		resp.PublishContext[KPublishSecondaryAttachId] = newAttachment.Id
	}

	return resp, nil
}

// ControllerUnpublishVolume implementation
func (c *Plugin) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	glog.V(5).Infof("start to controller unpublish volume, req volume id = %v, dode id = %v, controller unpublish secrets =%v",
		req.VolumeId, req.NodeId, req.Secrets)
	defer glog.V(5).Info("end to controller unpublish volume")

	if req.VolumeId == "" {
		msg := "volume ID must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if req.NodeId == "" {
		msg := "node ID must be provided"
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	//check volume is exist
	volSpec, errVol := c.Cli.GetVolume(req.VolumeId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist: %v", req.VolumeId, errVol)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := c.Cli.ListVolumeAttachments()
	if err != nil {
		msg := fmt.Sprintf("failed to list volume attachments: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	hostName := strings.Split(req.NodeId, ",")[0]

	var acts []*model.VolumeAttachmentSpec

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == req.VolumeId && attachSpec.Host == hostName {
			acts = append(acts, attachSpec)
			break
		}
	}

	if r := c.getReplicationByVolume(req.VolumeId); r != nil {
		for _, attachSpec := range attachments {
			if attachSpec.VolumeId == r.SecondaryVolumeId && attachSpec.Host == hostName {
				acts = append(acts, attachSpec)
				break
			}
		}
	}

	for _, act := range acts {
		if ok := UnpublishAttachmentList.isExist(act.Id); !ok {
			glog.Infof("add attachment id %s into unpublish attachment list", act.Id)
			UnpublishAttachmentList.Add(act)
			UnpublishAttachmentList.PrintList()
		}
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities implementation
func (c *Plugin) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes implementation
func (c *Plugin) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	glog.V(5).Info("start to list volumes")
	defer glog.V(5).Info("end to list volumes")

	// only support list all the volumes at present
	volumes, err := c.Cli.ListVolumes()
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
func (c *Plugin) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	glog.V(5).Info("start to get capacity")
	defer glog.V(5).Info("end to get capacity")

	pools, err := c.Cli.ListPools()
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
func (c *Plugin) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	glog.V(5).Info("start to controller get capabilities")
	defer glog.V(5).Info("end to controller get capabilities")

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
func (c *Plugin) FindSnapshot(req *model.VolumeSnapshotSpec) (bool, bool, *model.VolumeSnapshotSpec, error) {
	isExist := false
	snapshots, err := c.Cli.ListVolumeSnapshots()

	if err != nil {
		glog.Errorf("list volume snapshots failed: %v", err)

		return false, false, nil, err
	}

	for _, snapshot := range snapshots {
		if snapshot.Name == req.Name {
			isExist = true

			if (snapshot.VolumeId == req.VolumeId) && (snapshot.ProfileId == req.ProfileId) {
				glog.V(5).Info("snapshot already exists and is compatible")

				return true, true, snapshot, nil
			}
		}
	}

	return isExist, false, nil, nil
}

// CreateSnapshot implementation
func (c *Plugin) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {

	defer glog.V(5).Info("end to create snapshot")
	glog.V(5).Infof("start to create snapshot, name: %v, source volume id: %v, create snapshot secrets: %v, parameters: %v",
		req.Name, req.SourceVolumeId, req.Secrets, req.Parameters)

	if 0 == len(req.Name) {
		return nil, status.Error(codes.InvalidArgument, "snapshot name cannot be empty")
	}

	if 0 == len(req.SourceVolumeId) {
		return nil, status.Error(codes.InvalidArgument, "source volume ID cannot be empty")
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

	glog.Infof("create snapshot request body: %v", snapReq)
	var snapshot *model.VolumeSnapshotSpec
	isExist, isCompatible, findSnapshot, err := c.FindSnapshot(snapReq)

	if err != nil {
		return nil, err
	}

	if isExist {
		if isCompatible {
			snapshot = findSnapshot
		} else {
			return nil, status.Error(codes.AlreadyExists,
				"snapshot already exists but is incompatible")
		}
	} else {
		createSnapshot, err := c.Cli.CreateVolumeSnapshot(snapReq)
		if err != nil {
			glog.Error("failed to create volume snapshot: %v", err)
			return nil, err
		}

		snapshot = createSnapshot
	}

	glog.V(5).Infof("snapshot = %v", snapshot)
	creationTime, err := c.convertStringToPtypesTimestamp(snapshot.CreatedAt)
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

func (c *Plugin) convertStringToPtypesTimestamp(timeStr string) (*timestamp.Timestamp, error) {
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
func (c *Plugin) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {

	defer glog.V(5).Info("end to delete snapshot")
	glog.V(5).Infof("start to delete snapshot, snapshot id: %v, delete snapshot secrets: %v!",
		req.SnapshotId, req.Secrets)

	if 0 == len(req.SnapshotId) {
		return nil, status.Error(codes.InvalidArgument, "snapshot id cannot be empty")
	}

	err := c.Cli.DeleteVolumeSnapshot(req.SnapshotId, nil)

	if nil != err {
		msg := fmt.Sprintf("delete snapshot failed: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots implementation
func (c *Plugin) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {

	defer glog.V(5).Info("end to list snapshots")
	glog.V(5).Infof("start to list snapshots, MaxEntries: %v, StartingToken: %v, SourceVolumeId: %v, SnapshotId: %v!",
		req.MaxEntries, req.StartingToken, req.SourceVolumeId, req.SnapshotId)

	var opts map[string]string
	allSnapshots, err := c.Cli.ListVolumeSnapshots(opts)
	if nil != err {
		msg := fmt.Sprintf("failed to list snapshots: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
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

	glog.V(5).Infof("filter result=%v", filterResult)
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
			return nil, status.Error(codes.Aborted, "parsing the starting token failed")
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
		creationTime, err := c.convertStringToPtypesTimestamp(snapshot.CreatedAt)
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

func (c *Plugin) waitForVolStatusStable(volumeID string) (*model.VolumeSpec, error) {

	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(5 * time.Minute)

	defer ticker.Stop()
	validVolumeStatus := []string{"error", "error_deleting", "error_restoring", "error_extending", "available", "in-use"}

	for {
		select {
		case <-ticker.C:
			vol, err := c.Cli.GetVolume(volumeID)
			if err != nil {
				return nil, fmt.Errorf("get volume %s failed: %v", volumeID, err)
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
	glog.Infof("the list of attachments in the context is %s", attachmentIDList)
}

// UnpublishRoutine implementation
func (c *Plugin) UnpublishRoutine() {
	for {
		listLen := UnpublishAttachmentList.GetLen()
		if listLen > 0 {
			var next *list.Element
			for e := UnpublishAttachmentList.GetHead(); e != nil; e = next {
				next = e.Next()
				act := e.Value.(*model.VolumeAttachmentSpec)

				err := c.Cli.DeleteVolumeAttachment(act.Id, act)
				if err != nil {
					glog.Errorf("the volume %s failed to unpublish from node %s, error: %v.", act.VolumeId, act.Host, err)
				} else {
					c.waitVolumeAttachmentDeleted(act, e)
				}
				time.Sleep(10 * time.Second)
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func (c *Plugin) waitVolumeAttachmentDeleted(act *model.VolumeAttachmentSpec, e *list.Element) {
	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ticker.C:
			attachment, _ := c.Cli.GetVolumeAttachment(act.Id)
			if attachment != nil {
				glog.Errorf("waiting for the volume: %s successfully to unpublish to node: %s", act.VolumeId, act.Host)
			} else {
				glog.V(5).Infof("the volume: %s successfully to unpublish to node: %s", act.VolumeId, act.Host)
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
			glog.V(5).Infof("ISCSI initiator is %s", v)
			return v, nil
		}
	}

	return "", errors.New("no ISCSI initiators found")
}

func extractFCInitiatorFromNodeInfo(nodeInfo string) ([]string, error) {
	var wwpns []string
	for _, v := range strings.Split(nodeInfo, ",") {
		if strings.Contains(v, "node_name") {
			wwpns = append(wwpns, strings.Split(v, ":")[1])
		}
	}

	if len(wwpns) == 0 {
		return nil, errors.New("no FC initiators found")
	}

	glog.V(5).Infof("FC initiators are %s", wwpns)

	return wwpns, nil
}
