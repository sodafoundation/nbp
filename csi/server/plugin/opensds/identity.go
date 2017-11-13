package opensds

import (
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
)

////////////////////////////////////////////////////////////////////////////////
//                            Identity Service                                //
////////////////////////////////////////////////////////////////////////////////

var supportedVersions = []*csi.Version{
	&csi.Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
	},
	&csi.Version{
		Major: 1,
		Minor: 0,
		Patch: 0,
	},
}

// GetSupportedVersions implementation
func (p *Plugin) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	log.Println("start to GetSupportedVersions")
	defer log.Println("end to GetSupportedVersions")
	return &csi.GetSupportedVersionsResponse{
		SupportedVersions: supportedVersions,
	}, nil
}

// GetPluginInfo implementation
func (p *Plugin) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	log.Println("start to GetPluginInfo")
	defer log.Println("end to GetPluginInfo")

	return &csi.GetPluginInfoResponse{
		Name:          PluginName,
		VendorVersion: req.Version.String(),
		Manifest:      nil,
	}, nil
}
