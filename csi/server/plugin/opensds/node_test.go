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
	"reflect"
	"testing"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/opensds/nbp/client/iscsi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNodeGetId(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.NodeGetIdRequest{}
	iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}
	expectedNodeId := localIqn

	rs, err := fakePlugin.NodeGetId(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to GetNodeID: %v\n", err)
	}

	if rs.NodeId != expectedNodeId {
		t.Errorf("expected: %s, actual: %s\n", rs.NodeId, expectedNodeId)
	}
}

func TestNodeGetCapabilities(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.NodeGetCapabilitiesRequest{}
	expectedNodeCapabilities := []*csi.NodeServiceCapability{
		&csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
				},
			},
		},
	}

	rs, err := fakePlugin.NodeGetCapabilities(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to NodeGetCapabilities: %v\n", err)
	}

	if !reflect.DeepEqual(rs.Capabilities, expectedNodeCapabilities) {
		t.Errorf("expected: %v, actual: %v\n", rs.Capabilities, expectedNodeCapabilities)
	}
}

func TestNodeStageVolume(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := csi.NodeStageVolumeRequest{}

	_, err := fakePlugin.NodeStageVolume(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "Volume_id/staging_target_path/volume_capability must be specified")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}

	fakeReq.VolumeId = "bd5b12a8-a101-11e7-941e-d77981b584d9"
	fakeReq.StagingTargetPath = "123"
	fakeReq.VolumeCapability = &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{
				FsType:     "",
				MountFlags: []string{""},
			},
		},
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
	}

	fakeReq.PublishInfo = map[string]string{KPublishAttachId: "f2dda3d2-bf79-11e7-8665-f750b088f63e"}

	_, err = fakePlugin.NodeStageVolume(fakeCtx, &fakeReq)
	expectedErr = status.Error(codes.NotFound, "Volume does not exist")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}
}

func TestNodeUnstageVolume(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := csi.NodeUnstageVolumeRequest{}

	_, err := fakePlugin.NodeUnstageVolume(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "Volume_id/staging_target_path must be specified")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}
}

func TestNodePublishVolume(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := csi.NodePublishVolumeRequest{}

	_, err := fakePlugin.NodePublishVolume(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "Volume_id/staging_target_path/target_path/volume_capability must be specified")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}
}

func TestNodeUnpublishVolume(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := csi.NodeUnpublishVolumeRequest{}

	_, err := fakePlugin.NodeUnpublishVolume(fakeCtx, &fakeReq)
	expectedErr := status.Error(codes.InvalidArgument, "Volume_id/target_path must be specified")

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("expected: %v, actual: %v\n", expectedErr, err)
	}
}
