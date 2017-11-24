package opensds

import (
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
)

func TestValidateVolumeCapabilities(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.ValidateVolumeCapabilitiesRequest{
		Version:  supportedVersions[0],
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

func TestControllerProbe(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.ControllerProbeRequest{
		Version: supportedVersions[0],
	}
	expectedControllerProbe := &csi.ControllerProbeResponse{}

	rs, err := fakePlugin.ControllerProbe(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to ControllerProbe: %v\n", err)
	}

	if !reflect.DeepEqual(rs, expectedControllerProbe) {
		t.Errorf("expected: %v, actual: %v\n", rs, expectedControllerProbe)
	}
}

func TestControllerGetCapabilities(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.ControllerGetCapabilitiesRequest{
		Version: supportedVersions[0],
	}
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
	}

	rs, err := fakePlugin.ControllerGetCapabilities(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to ControllerGetCapabilities: %v\n", err)
	}

	if !reflect.DeepEqual(rs.Capabilities, expectedControllerCapabilities) {
		t.Errorf("expected: %v, actual: %v\n", rs.Capabilities, expectedControllerCapabilities)
	}
}
