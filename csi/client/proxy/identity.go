// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
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

package proxy

import (
	"log"

	"golang.org/x/net/context"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/opensds/nbp/csi/util"
)

// Identity define
type Identity struct {
	client csi.IdentityClient
}

////////////////////////////////////////////////////////////////////////////////
//                            Identity Client Init                            //
////////////////////////////////////////////////////////////////////////////////

// GetIdentity return Identity struct
func GetIdentity(csiEndpoint string) (client Identity, err error) {
	conn, err := util.GetCSIClientConn(csiEndpoint)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := csi.NewIdentityClient(conn)

	return Identity{client: c}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                            Identity Client Proxy                           //
////////////////////////////////////////////////////////////////////////////////

// GetPluginInfo proxy
func (c *Identity) GetPluginInfo(
	ctx context.Context) (*csi.GetPluginInfoResponse, error) {

	req := &csi.GetPluginInfoRequest{}

	rs, err := c.client.GetPluginInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// GetPluginCapabilities proxy
func (c *Identity) GetPluginCapabilities(
	ctx context.Context) (*csi.GetPluginCapabilitiesResponse, error) {

	req := &csi.GetPluginCapabilitiesRequest{}

	rs, err := c.client.GetPluginCapabilities(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// Probe proxy
func (c *Identity) Probe(
	ctx context.Context) (*csi.ProbeResponse, error) {

	req := &csi.ProbeRequest{}

	rs, err := c.client.Probe(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs, nil
}
