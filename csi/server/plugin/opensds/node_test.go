package opensds

import (
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/client/iscsi"
	"golang.org/x/net/context"
)

func TestGetNodeID(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.GetNodeIDRequest{
		Version: supportedVersions[0],
	}
	expectedNodeId := iscsi.GetHostIp()

	rs, err := fakePlugin.GetNodeID(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to GetNodeID: %v\n", err)
	}

	if rs.NodeId != expectedNodeId {
		t.Errorf("expected: %s, actual: %s\n", rs.NodeId, expectedNodeId)
	}
}

func TestNodeProbe(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.NodeProbeRequest{
		Version: supportedVersions[0],
	}
	expectedNodeProbe := &csi.NodeProbeResponse{}

	rs, err := fakePlugin.NodeProbe(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to NodeProbe: %v\n", err)
	}

	if !reflect.DeepEqual(rs, expectedNodeProbe) {
		t.Errorf("expected: %v, actual: %v\n", rs, expectedNodeProbe)
	}
}

func TestNodeGetCapabilities(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.NodeGetCapabilitiesRequest{
		Version: supportedVersions[0],
	}
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
