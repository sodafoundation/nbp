// Copyright 2018 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

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
func Probe(
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
func GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest,
	pluginName string) (
	*csi.GetPluginInfoResponse, error) {

	glog.Info("start to get plugin info")
	defer glog.Info("end to get plugin info")

	return &csi.GetPluginInfoResponse{
		Name:          pluginName,
		VendorVersion: "",
	}, nil
}

// GetPluginCapabilities implementation
func GetPluginCapabilities(
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
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_OFFLINE,
					},
				},
			},
		},
	}, nil
}
