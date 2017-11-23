package opensds

import (
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
)

func TestGetSupportedVersions(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.GetSupportedVersionsRequest{}
	rs, err := fakePlugin.GetSupportedVersions(fakeCtx, fakeReq)

	if err != nil {
		t.Errorf("failed to GetSupportedVersions: %v\n", err)
	}

	if !reflect.DeepEqual(rs.SupportedVersions, supportedVersions) {
		t.Errorf("expected: %v, actual: %v\n", rs.SupportedVersions, supportedVersions)
	}
}

func TestGetPluginInfo(t *testing.T) {
	var fakePlugin = &Plugin{}
	var fakeCtx = context.Background()
	fakeReq := &csi.GetPluginInfoRequest{
		Version: supportedVersions[0],
	}

	expectedPluginInfo := &csi.GetPluginInfoResponse{
		Name:          PluginName,
		VendorVersion: fakeReq.Version.String(),
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
