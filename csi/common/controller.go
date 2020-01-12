// Copyright 2019 The OpenSDS Authors.
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

package common

import (
	"container/list"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/opensds/nbp/csi/util"
	"github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	"github.com/opensds/opensds/pkg/utils/constants"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

// ValidateCreateVolReq - validates input paras of CreateVolume request
func ValidateCreateVolReq(req *csi.CreateVolumeRequest) error {

	if req.GetName() == "" {
		msg := "volume name must be provided when creating volume"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	if req.GetVolumeCapabilities() == nil || len(req.GetVolumeCapabilities()) == 0 {
		msg := "volume capabilities must be provided when creating volume"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	params := req.GetParameters()

	if params == nil {
		msg := "input parameters cannot be nil"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	keyList := []string{ParamProfile, ParamEnableReplication, ParamSecondaryAZ, PublishAttachMode}

	for k := range params {
		if !util.Contained(k, keyList) {
			msg := fmt.Sprintf("invalid input paramter key: %s. It should be one of %s,%s,%s,%s",
				k, ParamProfile, ParamEnableReplication, ParamSecondaryAZ, PublishAttachMode)
			glog.Error(msg)
			return status.Error(codes.InvalidArgument, msg)
		}
	}
	return nil
}

// ValidateDelVolReq - validates input paras of DeleteVolume request
func ValidateDelVolReq(req *csi.DeleteVolumeRequest) error {

	if req.VolumeId == "" {
		msg := "volume ID must be provided when deleting volume"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	return nil
}

// ValidateCtrlPubVolReq - validates input paras of ControllerPublishVolume request
func ValidateCtrlPubVolReq(req *csi.ControllerPublishVolumeRequest) error {

	if req.VolumeId == "" {
		msg := "volume ID must be provided when deleting volume"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	return nil
}

// ValidateCtrlUnPubVolReq - validates input paras of ControllerUnpublishVolume request
func ValidateCtrlUnPubVolReq(req *csi.ControllerUnpublishVolumeRequest) error {

	if req.VolumeId == "" {
		msg := "volume ID must be provided when deleting volume"
		glog.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	return nil
}

// ValidateVolumeCapabilities implementation
func ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity implementation
func GetCapacity(
	Client *client.Client,
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	glog.V(5).Info("start to get capacity")
	defer glog.V(5).Info("end to get capacity")

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
func ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	glog.V(5).Info("start to controller get capabilities")
	defer glog.V(5).Info("end to controller get capabilities")

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_GET_CAPACITY,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
	}, nil
}

// findSnapshot implementation
func findSnapshot(
	Client *client.Client,
	req *model.VolumeSnapshotSpec) (
	bool, bool, *model.VolumeSnapshotSpec, error) {
	isExist := false
	snapshots, err := Client.ListVolumeSnapshots()

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
func CreateSnapshot(
	Client *client.Client,
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
		case ParamProfile:
			if v == "" {
				msg := "profile id cannot be empty"
				glog.Error(msg)
				return nil, status.Error(codes.InvalidArgument, msg)
			}
			snapReq.ProfileId = v
		}
	}

	glog.Infof("create snapshot request body: %v", snapReq)
	var snapshot *model.VolumeSnapshotSpec
	isExist, isCompatible, findSnapshot, err := findSnapshot(Client, snapReq)

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
		createSnapshot, err := Client.CreateVolumeSnapshot(snapReq)
		if err != nil {
			glog.Errorf("failed to create volume snapshot: %v", err)
			return nil, err
		}

		snapshot = createSnapshot
	}

	glog.V(5).Infof("snapshot = %v", snapshot)
	creationTime, err := convertStringToPtypesTimestamp(snapshot.CreatedAt)
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

// convertStringToPtypesTimestamp converts to protobuf timestamp
func convertStringToPtypesTimestamp(timeStr string) (*timestamp.Timestamp, error) {
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
func DeleteSnapshot(
	Client *client.Client,
	ctx context.Context,
	req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {

	defer glog.V(5).Info("end to delete snapshot")
	glog.V(5).Infof("start to delete snapshot, snapshot id: %v, delete snapshot secrets: %v!",
		req.SnapshotId, req.Secrets)

	if 0 == len(req.SnapshotId) {
		return nil, status.Error(codes.InvalidArgument, "snapshot id cannot be empty")
	}

	err := Client.DeleteVolumeSnapshot(req.SnapshotId, nil)

	if nil != err {
		msg := fmt.Sprintf("delete snapshot failed: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots implementation
func ListSnapshots(
	Client *client.Client,
	ctx context.Context,
	req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {

	defer glog.V(5).Info("end to list snapshots")
	glog.V(5).Infof("start to list snapshots, MaxEntries: %v, StartingToken: %v, SourceVolumeId: %v, SnapshotId: %v!",
		req.MaxEntries, req.StartingToken, req.SourceVolumeId, req.SnapshotId)

	var opts map[string]string
	allSnapshots, err := Client.ListVolumeSnapshots(opts)
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
		creationTime, err := convertStringToPtypesTimestamp(snapshot.CreatedAt)
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

// AttachmentObj implementation
type AttachmentObj struct {
	l *list.List
	m sync.Mutex
	r sync.RWMutex
}

// NewList implementation
func NewList() *AttachmentObj {
	return &AttachmentObj{
		l: list.New(),
	}
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
func (q *AttachmentObj) IsExist(v interface{}) bool {
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

// PrintVolAttachList implementation
func (q *AttachmentObj) PrintVolAttachList() {
	var attachmentIDList string
	for e := q.GetHead(); e != nil; e = e.Next() {
		attachmentIDList = attachmentIDList + e.Value.(*model.VolumeAttachmentSpec).Id + ","
	}
	glog.Infof("the list of attachments in the context is %s", attachmentIDList)
}

// PrintFileShareList implementation
func (q *AttachmentObj) PrintFileShareList() {
	var attachmentIDList string
	for e := q.GetHead(); e != nil; e = e.Next() {
		attachmentIDList = attachmentIDList + e.Value.(*model.FileShareAclSpec).Id + ","
	}
	glog.Infof("the list of attachments in the context is %s", attachmentIDList)
}

// UnpublishRoutine implementation
func UnpublishRoutine(Client *client.Client) {
	UnpublishAttachmentList = NewList()
	for {
		listLen := UnpublishAttachmentList.GetLen()
		if listLen > 0 {
			var next *list.Element
			for e := UnpublishAttachmentList.GetHead(); e != nil; e = next {
				next = e.Next()

				switch e.Value.(type) {
				// delete volume attachment if storage type is block
				case *model.VolumeAttachmentSpec:
					act := e.Value.(*model.VolumeAttachmentSpec)

					if err := Client.DeleteVolumeAttachment(act.Id, act); err != nil {
						glog.Errorf("%s failed to unpublish: %v", act.Id, err)
					} else {
						waitAttachmentDeleted(act.Id, func(id string) (interface{}, error) {
							return Client.GetVolumeAttachment(id)
						}, e)
					}

				// delete fileshare access control list if storage type is file
				case *model.FileShareAclSpec:
					act := e.Value.(*model.FileShareAclSpec)

					if err := Client.DeleteFileShareAcl(act.Id); err != nil {
						if strings.Contains(err.Error(), "Not Found") {
							glog.Infof("delete attachment %s successfully", act.Id)
							UnpublishAttachmentList.Delete(e)
						} else {
							glog.Errorf("%s failed to unpublish: %v", act.Id, err)
						}
					} else {
						waitAttachmentDeleted(act.Id, func(id string) (interface{}, error) {
							return Client.GetFileShareAcl(id)
						}, e)
					}
				}

				time.Sleep(10 * time.Second)
			}
		}

		time.Sleep(10 * time.Second)
	}
}

// waitAttachmentDeleted waits for attachment deletion
func waitAttachmentDeleted(id string, f func(string) (interface{}, error), e *list.Element) {
	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ticker.C:
			_, err := f(id)

			if err != nil && strings.Contains(err.Error(), "Not Found") {
				glog.Infof("delete attachment %s successfully", id)
				UnpublishAttachmentList.Delete(e)
				return
			} else {
				glog.Errorf("delete attachment failed: %v", err)
			}

		case <-timeout:
			glog.Errorf("waiting to delete %s timeout", id)
			return
		}
	}
}
