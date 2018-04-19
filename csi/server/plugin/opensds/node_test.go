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
