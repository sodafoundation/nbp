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

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
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
func GetNode(csiEndpoint string) (client Node, err error) {
	conn, err := util.GetCSIClientConn(csiEndpoint)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := csi.NewNodeClient(conn)

	return Node{client: c}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                            Node Client Proxy                               //
////////////////////////////////////////////////////////////////////////////////

// NodeStageVolume proxy
func (c *Node) NodeStageVolume(
	ctx context.Context,
	volumeid string,
	publicInfo map[string]string, /*Optional*/
	stagingTargetPath string,
	capability *csi.VolumeCapability,
	credentials map[string]string, /*Optional*/
	volumeattributes map[string]string /*Optional*/) error {

	req := &csi.NodeStageVolumeRequest{
		VolumeId:          volumeid,
		PublishInfo:       publicInfo,
		StagingTargetPath: stagingTargetPath,
		VolumeCapability:  capability,
		NodeStageSecrets:  credentials,
		VolumeAttributes:  volumeattributes,
	}

	_, err := c.client.NodeStageVolume(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// NodeUnstageVolume proxy
func (c *Node) NodeUnstageVolume(
	ctx context.Context,
	volumeid string,
	stagingTargetPath string) error {

	req := &csi.NodeUnstageVolumeRequest{
		VolumeId:          volumeid,
		StagingTargetPath: stagingTargetPath,
	}

	_, err := c.client.NodeUnstageVolume(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// NodePublishVolume proxy
func (c *Node) NodePublishVolume(
	ctx context.Context,
	volumeid string,
	publicInfo map[string]string, /*Optional*/
	stagingTargetPath string, /*Optional*/
	targetPath string,
	capability *csi.VolumeCapability,
	readonly bool,
	credentials map[string]string, /*Optional*/
	volumeattributes map[string]string /*Optional*/) error {

	req := &csi.NodePublishVolumeRequest{
		VolumeId:           volumeid,
		PublishInfo:        publicInfo,
		StagingTargetPath:  stagingTargetPath,
		TargetPath:         targetPath,
		VolumeCapability:   capability,
		Readonly:           readonly,
		NodePublishSecrets: credentials,
		VolumeAttributes:   volumeattributes,
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
	volumeid string,
	targetPath string) error {

	req := &csi.NodeUnpublishVolumeRequest{
		VolumeId:   volumeid,
		TargetPath: targetPath,
	}

	_, err := c.client.NodeUnpublishVolume(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// NodeGetId proxy
func (c *Node) NodeGetId(
	ctx context.Context) (string, error) {

	req := &csi.NodeGetIdRequest{}

	rs, err := c.client.NodeGetId(ctx, req)
	if err != nil {
		return "", err
	}

	return rs.NodeId, nil
}

// NodeGetCapabilities proxy
func (c *Node) NodeGetCapabilities(
	ctx context.Context) (capabilties []*csi.NodeServiceCapability, err error) {

	req := &csi.NodeGetCapabilitiesRequest{}

	rs, err := c.client.NodeGetCapabilities(ctx, req)
	if err != nil {
		return nil, err
	}

	return rs.Capabilities, nil
}
