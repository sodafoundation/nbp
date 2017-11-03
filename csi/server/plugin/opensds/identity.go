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

//CheckVersionSupport check whether api version is supported
func (p *Plugin) CheckVersionSupport(version *csi.Version) (bool, csi.Error_GeneralError_GeneralErrorCode) {
	if version == nil {
		return false, csi.Error_GeneralError_UNDEFINED
	}

	for _, ver := range supportedVersions {
		if version == ver {
			return true, csi.Error_GeneralError_UNKNOWN
		}
	}

	return false, csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION
}

// GetSupportedVersions implementation
func (p *Plugin) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	log.Println("start to GetSupportedVersions")
	defer log.Println("end to GetSupportedVersions")
	return &csi.GetSupportedVersionsResponse{
		Reply: &csi.GetSupportedVersionsResponse_Result_{
			Result: &csi.GetSupportedVersionsResponse_Result{
				SupportedVersions: supportedVersions,
			},
		},
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
		Reply: &csi.GetPluginInfoResponse_Result_{
			Result: &csi.GetPluginInfoResponse_Result{
				Name:          PluginName,
				VendorVersion: req.Version.String(),
				Manifest:      nil,
			},
		},
	}, nil
}
