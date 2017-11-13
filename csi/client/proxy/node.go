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
	volumeid string,
	volumeinfo map[string]string, /*Optional*/
	targetPath string,
	capability *csi.VolumeCapability,
	readonly bool,
	credentials map[string]string, /*Optional*/
	volumeattributes map[string]string /*Optional*/) error {

	req := &csi.NodePublishVolumeRequest{
		Version:           version,
		VolumeId:          volumeid,
		PublishVolumeInfo: volumeinfo,
		TargetPath:        targetPath,
		VolumeCapability:  capability,
		Readonly:          readonly,
		UserCredentials:   credentials,
		VolumeAttributes:  volumeattributes,
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
	volumeid string,
	targetPath string,
	credentials map[string]string /*Optional*/) error {

	req := &csi.NodeUnpublishVolumeRequest{
		Version:         version,
		VolumeId:        volumeid,
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
	version *csi.Version) (string, error) {

	req := &csi.GetNodeIDRequest{
		Version: version,
	}

	rs, err := c.client.GetNodeID(ctx, req)
	if err != nil {
		return "", err
	}

	return rs.NodeId, nil
}

// NodeProbe proxy
func (c *Node) NodeProbe(
	ctx context.Context,
	version *csi.Version) error {

	req := &csi.NodeProbeRequest{
		Version: version,
	}

	_, err := c.client.NodeProbe(ctx, req)
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

	return rs.Capabilities, nil
}
