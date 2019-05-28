package opensds

import (
	"fmt"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	sdscontroller "github.com/opensds/nbp/client/opensds"
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

	glog.Info("start to probe")
	defer glog.Info("end to probe")

	_, err := sdscontroller.GetClient("", "")
	if err != nil {
		msg := fmt.Sprintf("failed to communicate with opensds client, %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	return &csi.ProbeResponse{}, nil
}

// GetPluginInfo implementation
func (p *Plugin) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	glog.Info("start to get plugin info")
	defer glog.Info("end to get plugin info")

	return &csi.GetPluginInfoResponse{
		Name:          PluginName,
		VendorVersion: "",
	}, nil
}

// GetPluginCapabilities implementation
func (p *Plugin) GetPluginCapabilities(
	ctx context.Context,
	req *csi.GetPluginCapabilitiesRequest) (
	*csi.GetPluginCapabilitiesResponse, error) {

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
		},
	}, nil
}
