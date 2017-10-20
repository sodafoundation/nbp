package main

import (
	"context"
	"log"

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
}
