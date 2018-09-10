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

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	c "github.com/opensds/opensds/client"

	"github.com/opensds/opensds/pkg/model"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	Client.VolumeMgr = fv
}

var fv = &c.VolumeMgr{
	Receiver: NewFakeVolumeReceiver(),
}
var (
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
			"volumeId": "bd5b12a8-a101-11e7-941e-d77981b584d8"
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
		case *model.VolumeSnapshotSpec:
			if err := json.Unmarshal([]byte(ByteSnapshot), out); err != nil {
				return err
			}
			break
		}
	case "GET":
		switch out.(type) {
		case *[]*model.VolumeSnapshotSpec:
			if err := json.Unmarshal([]byte(ByteSnapshots), out); err != nil {
				return err
			}
			break
		}
	case "DELETE":
		break
	default:
		return errors.New("inputed method format not supported!")
	}

	return nil
}

func TestValidateVolumeCapabilities(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId: "1234567890",
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
		VolumeAttributes: map[string]string{"key": "value"},
	}
	expectedValidateVolumeCapabilities := &csi.ValidateVolumeCapabilitiesResponse{
		Supported: true,
		Message:   "supported",
	}

	rs, err := fakePlugin.ValidateVolumeCapabilities(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to ValidateVolumeCapabilities: %v\n", err)
	}

	if !reflect.DeepEqual(rs, expectedValidateVolumeCapabilities) {
		t.Errorf("expected: %v, actual: %v\n", rs, expectedValidateVolumeCapabilities)
	}
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
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := csi.CreateSnapshotRequest{}

	rs, err := fakePlugin.CreateSnapshot(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "Snapshot Name cannot be empty")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}

	fakeReq.Name = "volume00"
	rs, err = fakePlugin.CreateSnapshot(fakeCtx, &fakeReq)
	expectedErr = status.Error(codes.InvalidArgument, "Source Volume ID cannot be empty")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}

	fakeReq.SourceVolumeId = "b5e56f11-ea23-4aa0-b6f3-f902d4892bbb"
	rs, err = fakePlugin.CreateSnapshot(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to CreateSnapshot: %v\n", err)
	}

	expectedResponse := csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      1 * 1024 * 1024 * 1024,
			Id:             "3769855c-a102-11e7-b772-17b880d2f537",
			SourceVolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
			CreatedAt:      1536167248000000000,
			Status: &csi.SnapshotStatus{
				Type: csi.SnapshotStatus_READY,
			},
		},
	}

	if !reflect.DeepEqual(&expectedResponse, rs) {
		t.Errorf("expected: %v, actual: %v\n", &expectedResponse, rs)
	}
}

func TestDeleteSnapshot(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := csi.DeleteSnapshotRequest{}

	rs, err := fakePlugin.DeleteSnapshot(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "Snapshot ID cannot be empty")

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
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := csi.ListSnapshotsRequest{}

	rs, err := fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	fakeReq.SnapshotId = "3769855c-a102-11e7-b772-17b880d2f537"
	rs, err = fakePlugin.ListSnapshots(fakeCtx, &fakeReq)

	if nil != err {
		t.Errorf("failed to ListSnapshots: %v\n", err)
	}

	expectedEntries := []*csi.ListSnapshotsResponse_Entry{
		&csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SizeBytes:      1 * 1024 * 1024 * 1024,
				Id:             "3769855c-a102-11e7-b772-17b880d2f537",
				SourceVolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
				CreatedAt:      1536167248000000000,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		},
		&csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SizeBytes:      1 * 1024 * 1024 * 1024,
				Id:             "3bfaf2cc-a102-11e7-8ecb-63aea739d755",
				SourceVolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
				CreatedAt:      1536167248000000000,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		},
	}

	expectedRs := &csi.ListSnapshotsResponse{Entries: expectedEntries}

	if !reflect.DeepEqual(expectedRs, rs) {
		t.Errorf("expected: %v, actual: %v\n", expectedEntries, rs)
	}
}
