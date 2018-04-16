package opensds

import (
	"reflect"
	"testing"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"golang.org/x/net/context"
)

func TestGetPluginInfo(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.GetPluginInfoRequest{}

	expectedPluginInfo := &csi.GetPluginInfoResponse{
		Name:          PluginName,
		VendorVersion: "",
		Manifest:      nil,
	}

	rs, err := fakePlugin.GetPluginInfo(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to GetPluginInfo: %v\n", err)
	}

	if !reflect.DeepEqual(rs, expectedPluginInfo) {
		t.Errorf("expected: %v, actual: %v\n", rs, expectedPluginInfo)
	}
}
