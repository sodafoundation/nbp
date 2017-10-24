package main

import (
	"context"
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/csi/client/proxy"
)

func main() {
	// GetIdentity
	log.Println("start to get identity")
	identity, err := proxy.GetIdentity()
	if err != nil {
		log.Fatalf("failed to get identity: %v", err)
	}

	// Test GetSupportedVersions
	versions, err := identity.GetSupportedVersions(context.Background())
	if err != nil {
		log.Fatalf("failed to GetSupportedVersions: %v", err)
	}

	// printf results
	for _, v := range versions {
		log.Printf("[GetSupportedVersionsResponse] version:%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	if len(versions) == 0 {
		log.Print("versions length is 0")
		return
	}

	// Test GetPluginInfo
	rs, err := identity.GetPluginInfo(context.Background(), versions[0])
	if err != nil {
		log.Fatalf("failed to GetPluginInfo: %v", err)
	}

	// printf results
	log.Printf("[GetPluginInfoResponse] Name:%s VendorVersion:%s Manifest:%v",
		rs.Name, rs.VendorVersion, rs.Manifest)

	// GetController
	log.Println("start to get controller")
	controller, err := proxy.GetController()
	if err != nil {
		log.Fatalf("failed to get controller: %v", err)
	}

	// Test ControllerGetCapabilities
	controllercapabilities, err := controller.ControllerGetCapabilities(context.Background(), versions[0])
	if err != nil {
		log.Fatalf("failed to ControllerGetCapabilities: %v", err)
	}

	log.Printf("[ControllerGetCapabilities] controllercapabilities:%v", controllercapabilities)

	// Test ValidateVolumeCapabilities
	volumeinfo := &csi.VolumeInfo{
		Handle: &csi.VolumeHandle{
			Id: "1234567890",
			Metadata: map[string]string{
				"key": "value",
			},
		},
		CapacityBytes: 1000,
	}

	volumecapabilities := []*csi.VolumeCapability{
		&csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
		},
	}

	validatevolumecapabilities, err := controller.ValidateVolumeCapabilities(context.Background(), versions[0], volumeinfo, volumecapabilities)
	if err != nil {
		log.Fatalf("failed to ValidateVolumeCapabilities: %v", err)
	}

	log.Printf("[ValidateVolumeCapabilities] validatevolumecapabilities:%v", validatevolumecapabilities)

	// GetNode
	log.Println("start to get node")
	node, err := proxy.GetNode()
	if err != nil {
		log.Fatalf("failed to get node: %v", err)
	}

	// Test NodeGetCapabilities
	nodecapabilities, err := node.NodeGetCapabilities(context.Background(), versions[0])
	if err != nil {
		log.Fatalf("failed to NodeGetCapabilities: %v", err)
	}

	log.Printf("[NodeGetCapabilities] nodecapabilities:%v", nodecapabilities)

	// Test GetNodeID
	nodeid, err := node.GetNodeID(context.Background(), versions[0])
	if err != nil {
		log.Fatalf("failed to GetNodeID: %v", err)
	}

	log.Printf("[GetNodeID] nodeid:%v", nodeid)

	// Test ProbeNode
	err = node.ProbeNode(context.Background(), versions[0])
	if err != nil {
		log.Fatalf("failed to ProbeNode: %v", err)
	} else {
		log.Printf("[ProbeNode] ProbeNode:OK")
	}
}
