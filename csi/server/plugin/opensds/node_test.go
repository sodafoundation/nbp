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
					Type: csi.NodeServiceCapability_RPC_UNKNOWN,
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
