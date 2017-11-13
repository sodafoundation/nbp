package proxy

import (
	"log"

	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
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
func GetIdentity() (client Identity, err error) {
	conn, err := util.GetCSIClientConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := csi.NewIdentityClient(conn)

	return Identity{client: c}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                            Identity Client Proxy                           //
////////////////////////////////////////////////////////////////////////////////

// GetSupportedVersions proxy
func (c *Identity) GetSupportedVersions(
	ctx context.Context) ([]*csi.Version, error) {

	req := &csi.GetSupportedVersionsRequest{}

	rs, err := c.client.GetSupportedVersions(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs.SupportedVersions, nil
}

// GetPluginInfo proxy
func (c *Identity) GetPluginInfo(
	ctx context.Context,
	version *csi.Version) (*csi.GetPluginInfoResponse, error) {

	req := &csi.GetPluginInfoRequest{
		Version: version,
	}

	rs, err := c.client.GetPluginInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs, nil
}
