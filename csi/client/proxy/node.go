package proxy

import (
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/csi/util"
	"golang.org/x/net/context"
)

// Node define
type Node struct {
	client csi.NodeClient
}

////////////////////////////////////////////////////////////////////////////////
//                            Node Client Init                                //
////////////////////////////////////////////////////////////////////////////////

// GetNode return Node struct
func GetNode() (client Node, err error) {
	conn, err := util.GetCSIClientConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := csi.NewNodeClient(conn)

	return Node{client: c}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                            Node Client Proxy                               //
////////////////////////////////////////////////////////////////////////////////

// NodePublishVolume proxy
func (c *Node) NodePublishVolume(
	ctx context.Context,
	version *csi.Version,
	handle *csi.VolumeHandle,
	volumeinfo map[string]string, /*Optional*/
	targetPath string,
	capability *csi.VolumeCapability,
	readonly bool,
	credentials *csi.Credentials /*Optional*/) error {

	req := &csi.NodePublishVolumeRequest{
		Version:           version,
		VolumeHandle:      handle,
		PublishVolumeInfo: volumeinfo,
		TargetPath:        targetPath,
		VolumeCapability:  capability,
		Readonly:          readonly,
		UserCredentials:   credentials,
	}

	_, err := c.client.NodePublishVolume(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// NodeUnpublishVolume proxy
func (c *Node) NodeUnpublishVolume(
	ctx context.Context,
	version *csi.Version,
	handle *csi.VolumeHandle,
	targetPath string,
	credentials *csi.Credentials /*Optional*/) error {

	req := &csi.NodeUnpublishVolumeRequest{
		Version:         version,
		VolumeHandle:    handle,
		TargetPath:      targetPath,
		UserCredentials: credentials,
	}

	_, err := c.client.NodeUnpublishVolume(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetNodeID proxy
func (c *Node) GetNodeID(
	ctx context.Context,
	version *csi.Version) (*csi.NodeID, error) {

	req := &csi.GetNodeIDRequest{
		Version: version,
	}

	rs, err := c.client.GetNodeID(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs.GetResult().NodeId, nil
}

// ProbeNode proxy
func (c *Node) ProbeNode(
	ctx context.Context,
	version *csi.Version) error {

	req := &csi.ProbeNodeRequest{
		Version: version,
	}

	_, err := c.client.ProbeNode(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// NodeGetCapabilities proxy
func (c *Node) NodeGetCapabilities(
	ctx context.Context,
	version *csi.Version) (capabilties []*csi.NodeServiceCapability, err error) {

	req := &csi.NodeGetCapabilitiesRequest{
		Version: version,
	}

	rs, err := c.client.NodeGetCapabilities(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs.GetResult().Capabilities, nil
}
