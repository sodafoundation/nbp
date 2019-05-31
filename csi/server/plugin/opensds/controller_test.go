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
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/csi/util"
	"github.com/opensds/opensds/client"
	c "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var fakePlugin *Plugin
var fakeCtx context.Context

func init() {
	client := &client.Client{}

	client.VolumeMgr = &c.VolumeMgr{
		Receiver: NewFakeVolumeReceiver(),
	}

	client.ReplicationMgr = &c.ReplicationMgr{
		Receiver: NewFakeReplicationReceiver(),
		Endpoint: "0.0.0.0",
		TenantId: "123456",
	}
	client.ProfileMgr = &c.ProfileMgr{
		Receiver: NewFakeProfileReceiver(),
	}
	fakePlugin = &Plugin{Client: client}
	fakeCtx = context.Background()
}

func NewFakeReplicationReceiver() c.Receiver {
	return &fakeReplicationReceiver{}
}

type fakeReplicationReceiver struct{}

func (*fakeReplicationReceiver) Recv(
	string,
	method string,
	in interface{},
	out interface{},
) error {
	return nil
}

var (
	ByteVolume = `{
		"id": "bd5b12a8-a101-11e7-941e-d77981b584d8",
		"name": "sample-volume",
		"description": "This is a sample volume for testing",
		"size": 1,
		"availabilityZone": "default",
		"status": "available",
		"poolId": "084bf71e-a102-11e7-88a8-e31fe6d52248",
		"profileId": "1106b972-66ef-11e7-b172-db03f3689c9c"
	}`

	ByteProfile = `{
	
			"id": "1106b972-66ef-11e7-b172-db03f3689c9c",
			"name": "default",
			"description": "default policy",
			"storageType": "block"
		
}`

	ByteVolumes = `[
		{
			"id": "bd5b12a8-a101-11e7-941e-d77981b584d8",
			"name": "sample-volume-1",
			"description": "This is a sample volume for testing",
			"size": 1,
			"status": "available",
			"poolId": "084bf71e-a102-11e7-88a8-e31fe6d52248",
			"profileId": "1106b972-66ef-11e7-b172-db03f3689c9c"
		}
	]`

	ByteSnapshot = `{
		"id": "3769855c-a102-11e7-b772-17b880d2f537",
		"createdAt":"2018-09-05T17:07:28",
		"name": "sample-snapshot-01",
		"description": "This is the first sample snapshot for testing",
		"size": 1,
		"status": "available",
		"volumeId": "bd5b12a8-a101-11e7-941e-d77981b584d8"
	}`

	ByteSnapshots = `[
		{
			"id": "3769855c-a102-11e7-b772-17b880d2f537",
			"createdAt":"2018-09-05T17:07:28",
			"name": "sample-snapshot-01",
			"description": "This is the first sample snapshot for testing",
			"size": 1,
			"status": "available",
			"volumeId": "bd5b12a8-a101-11e7-941e-d77981b584d8"
		},
		{
			"id": "3bfaf2cc-a102-11e7-8ecb-63aea739d755",
			"createdAt":"2018-09-05T17:07:28",
			"name": "sample-snapshot-02",
			"description": "This is the second sample snapshot for testing",
			"size": 1,
			"status": "available",
			"volumeId": "bd5b12a8-a101-11e7-941e-d77981b584d9"
		}
	]`
)

func NewFakeVolumeReceiver() c.Receiver {
	return &fakeVolumeReceiver{}
}

type fakeVolumeReceiver struct{}

func (*fakeVolumeReceiver) Recv(
	string,
	method string,
	in interface{},
	out interface{},
) error {
	switch strings.ToUpper(method) {
	case "POST", "PUT":
		switch out.(type) {
		case *model.VolumeSpec:
			if err := json.Unmarshal([]byte(ByteVolume), out); err != nil {
				return err
			}
			break
		case *model.VolumeSnapshotSpec:
			if err := json.Unmarshal([]byte(ByteSnapshot), out); err != nil {
				return err
			}
			break
		default:
			return errors.New("output format not supported")
		}
		break
	case "GET":
		switch out.(type) {
		case *[]*model.VolumeSnapshotSpec:
			if err := json.Unmarshal([]byte(ByteSnapshots), out); err != nil {
				return err
			}
			break
		case *[]*model.VolumeSpec:
			if err := json.Unmarshal([]byte(ByteVolumes), out); err != nil {
				return err
			}
			break
		case *model.VolumeSpec:
			if err := json.Unmarshal([]byte(ByteVolume), out); err != nil {
				return err
			}
			break
		default:
			return errors.New("output format not supported")
		}
		break
	case "DELETE":
		break
	default:
		return errors.New("inputed method format not supported")
	}

	return nil
}

func NewFakeProfileReceiver() c.Receiver {
	return &fakeProfileReceiver{}
}

type fakeProfileReceiver struct{}

func (*fakeProfileReceiver) Recv(string, method string, in, out interface{}) error {
	if strings.ToUpper(method) == "GET" {
		if err := json.Unmarshal([]byte(ByteProfile), out); err != nil {
			return err
		}
	}
	return nil
}

func TestControllerGetCapabilities(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.ControllerGetCapabilitiesRequest{}
	expectedControllerCapabilities := []*csi.ControllerServiceCapability{
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
	}

	rs, err := fakePlugin.ControllerGetCapabilities(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to ControllerGetCapabilities: %v\n", err)
	}

	if !reflect.DeepEqual(rs.Capabilities, expectedControllerCapabilities) {
		t.Errorf("expected: %v, actual: %v\n", rs.Capabilities, expectedControllerCapabilities)
	}
}

func TestCreateSnapshot(t *testing.T) {
	fakeReq := csi.CreateSnapshotRequest{}

	rs, err := fakePlugin.CreateSnapshot(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "snapshot name cannot be empty")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}

	fakeReq.Name = "volume00"
	rs, err = fakePlugin.CreateSnapshot(fakeCtx, &fakeReq)
	expectedErr = status.Error(codes.InvalidArgument, "source volume ID cannot be empty")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}

	fakeReq.SourceVolumeId = "b5e56f11-ea23-4aa0-b6f3-f902d4892bbb"
	rs, err = fakePlugin.CreateSnapshot(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to CreateSnapshot: %v\n", err)
	}

	ptypesTime, err := ptypes.TimestampProto(time.Unix(0, 1536167248000000000))
	if err != nil {
		t.Errorf("failed to CreateSnapshot: %v\n", err)
	}

	expectedResponse := csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      util.GiB,
			SnapshotId:     "3769855c-a102-11e7-b772-17b880d2f537",
			SourceVolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
			CreationTime:   ptypesTime,
			ReadyToUse:     true,
		},
	}

	if !reflect.DeepEqual(&expectedResponse, rs) {
		t.Errorf("expected: %v, actual: %v\n", &expectedResponse, rs)
	}
}

func TestDeleteSnapshot(t *testing.T) {
	fakeReq := csi.DeleteSnapshotRequest{}

	rs, err := fakePlugin.DeleteSnapshot(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "snapshot id cannot be empty")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}

	fakeReq.SnapshotId = "3769855c-a102-11e7-b772-17b880d2f537"
	rs, err = fakePlugin.DeleteSnapshot(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to DeleteSnapshot: %v\n", err)
	}

	expectedResponse := &csi.DeleteSnapshotResponse{}

	if !reflect.DeepEqual(expectedResponse, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedResponse, rs)
	}
}

func TestListSnapshots(t *testing.T) {
	ptypesTime, err := ptypes.TimestampProto(time.Unix(0, 1536167248000000000))
	if err != nil {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}
	expectedEntries := []*csi.ListSnapshotsResponse_Entry{
		&csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SizeBytes:      util.GiB,
				SnapshotId:     "3769855c-a102-11e7-b772-17b880d2f537",
				SourceVolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
				CreationTime:   ptypesTime,
				ReadyToUse:     true,
			},
		},
		&csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SizeBytes:      util.GiB,
				SnapshotId:     "3bfaf2cc-a102-11e7-8ecb-63aea739d755",
				SourceVolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d9",
				CreationTime:   ptypesTime,
				ReadyToUse:     true,
			},
		},
	}

	expectedRs := &csi.ListSnapshotsResponse{Entries: expectedEntries}
	// 1、ListSnapshotsRequest no parameters
	fakeReq := csi.ListSnapshotsRequest{}
	rs, err := fakePlugin.ListSnapshots(fakeCtx, &fakeReq)
	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}

	// 2、ListSnapshotsRequest use only "SnapshotId" as a filter
	fakeReq.SnapshotId = "3769855c-a102-11e7-b772-17b880d2f537"
	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	expectedRs = &csi.ListSnapshotsResponse{Entries: expectedEntries[:1]}
	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}

	// 3、ListSnapshotsRequest use only "SourceVolumeId" as a filter
	fakeReq.SnapshotId = ""
	fakeReq.SourceVolumeId = "bd5b12a8-a101-11e7-941e-d77981b584d9"
	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	expectedRs = &csi.ListSnapshotsResponse{Entries: expectedEntries[1:2]}
	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}

	// 4、ListSnapshotsRequest use "SourceVolumeId" and "SnapshotId"
	fakeReq.SnapshotId = "3769855c-a102-11e7-b772-17b880d2f537"
	fakeReq.SourceVolumeId = "bd5b12a8-a101-11e7-941e-d77981b584d8"
	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	expectedRs = &csi.ListSnapshotsResponse{Entries: expectedEntries[0:1]}
	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}

	// 5、ListSnapshotsRequest use "MaxEntries" and "StartingToken"
	fakeReq.SnapshotId = ""
	fakeReq.SourceVolumeId = ""
	fakeReq.MaxEntries = 2
	fakeReq.StartingToken = "1"

	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	expectedRs = &csi.ListSnapshotsResponse{Entries: expectedEntries[1:2]}
	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}

	fakeReq.MaxEntries = 3
	fakeReq.StartingToken = "0"

	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	expectedRs = &csi.ListSnapshotsResponse{Entries: expectedEntries}
	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}

	fakeReq.MaxEntries = 1
	fakeReq.StartingToken = "0"

	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	expectedRs = &csi.ListSnapshotsResponse{Entries: expectedEntries[0:1],
		NextToken: "1"}
	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}

	// Test error
	fakeReq.MaxEntries = 1
	fakeReq.StartingToken = "2"
	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.Aborted,
		"startingToken=2 >= len(snapshots)=2")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}

	fakeReq.MaxEntries = 1
	fakeReq.StartingToken = "k"
	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)
	expectedErr = status.Error(codes.Aborted, "parsing the starting token failed")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}
}

func TestCreateVolume(t *testing.T) {
	fakeReq := csi.CreateVolumeRequest{
		Name: "sample-volume",
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
				AccessType: &csi.VolumeCapability_Block{
					Block: &csi.VolumeCapability_BlockVolume{},
				},
			},
		},
		Parameters: map[string]string{
			"profile":          "1106b972-66ef-11e7-b172-db03f3689c9c",
			"availabilityzone": "default",
			"storageType":      "block",
		},
		VolumeContentSource: &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Snapshot{
				Snapshot: &csi.VolumeContentSource_SnapshotSource{
					SnapshotId: "3769855c-a102-11e7-b772-17b880d2f537",
				},
			},
		},
	}

	rs, err := fakePlugin.CreateVolume(fakeCtx, &fakeReq)
	if nil != err {
		t.Errorf("failed to CreateVolume: %v\n", err)
	}

	expectedVolumeinfo := &csi.Volume{
		CapacityBytes: util.GiB,
		VolumeId:      "bd5b12a8-a101-11e7-941e-d77981b584d8",
		VolumeContext: map[string]string{
			KVolumeName:        "sample-volume",
			KVolumeStatus:      "available",
			KVolumeAZ:          "default",
			KVolumePoolId:      "084bf71e-a102-11e7-88a8-e31fe6d52248",
			KVolumeProfileId:   "1106b972-66ef-11e7-b172-db03f3689c9c",
			KVolumeLvPath:      "",
			KPublishAttachMode: "rw",
		},
		AccessibleTopology: []*csi.Topology{
			{
				Segments: map[string]string{
					TopologyZoneKey: "default",
				},
			},
		},
	}

	expectedRs := &csi.CreateVolumeResponse{
		Volume: expectedVolumeinfo,
	}

	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedRs, rs)
	}
}

func TestIsStringMapEqual(t *testing.T) {
	metadataA := map[string]string{"lvPath": "/dev/opensds-volumes-default/volume-105a8e15-8ab2-463c-9efb-7af1a3451138"}
	metadataB := map[string]string{"lvPath": "/dev/opensds-volumes-default/volume-105a8e15-8ab2-463c-9efb-7af1a3451138"}
	ret := isStringMapEqual(metadataA, metadataB)

	if !ret {
		t.Errorf("expected: true, actual: %v\n", ret)
	}
}
