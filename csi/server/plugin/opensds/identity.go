package opensds

import (
	"runtime"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Identity Service                                //
////////////////////////////////////////////////////////////////////////////////

// Probe implementation
func (p *Plugin) Probe(
	ctx context.Context,
	req *csi.ProbeRequest) (
	*csi.ProbeResponse, error) {

	glog.Info("start to Probe")
	defer glog.Info("end to Probe")

	switch runtime.GOOS {
	case "linux":
		return &csi.ProbeResponse{}, nil
	default:
		msg := "unsupported operating system:" + runtime.GOOS
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}
}

// GetPluginInfo implementation
func (p *Plugin) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	glog.Info("start to GetPluginInfo")
	defer glog.Info("end to GetPluginInfo")

	return &csi.GetPluginInfoResponse{
		Name:          PluginName,
		VendorVersion: "",
		Manifest:      nil,
	}, nil
}

// GetPluginInfo implementation
func (p *Plugin) GetPluginCapabilities(
	ctx context.Context,
	req *csi.GetPluginCapabilitiesRequest) (
	*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			&csi.PluginCapability{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}
