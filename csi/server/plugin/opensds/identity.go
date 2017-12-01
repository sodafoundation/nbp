package opensds

import (
	"log"
	"reflect"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
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

//CheckVersionSupport check whether api version is supported
func (p *Plugin) CheckVersionSupport(version *csi.Version) codes.Code {
	if version == nil {
		return codes.InvalidArgument
	}

	for _, ver := range supportedVersions {
		if reflect.DeepEqual(version, ver) {
			return codes.OK
		}
	}

	return codes.InvalidArgument
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
