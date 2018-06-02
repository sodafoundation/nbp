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
