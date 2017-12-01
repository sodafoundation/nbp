package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

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

	// Test ControllerProbe
	err = controller.ControllerProbe(context.Background(), versions[0])
	if err != nil {
		log.Fatalf("failed to ControllerProbe: %v", err)
	} else {
		log.Printf("[ControllerProbe] ControllerProbe:OK")
	}

	// Test ValidateVolumeCapabilities
	volumeid := "1234567890"
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
	volumeattributes := map[string]string{
		"key": "value",
	}

	validatevolumecapabilities, err := controller.ValidateVolumeCapabilities(
		context.Background(),
		versions[0],
		volumeid,
		volumecapabilities,
		volumeattributes)
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

	// Test NodeProbe
	err = node.NodeProbe(context.Background(), versions[0])
	if err != nil {
		log.Fatalf("failed to NodeProbe: %v", err)
	} else {
		log.Printf("[NodeProbe] NodeProbe:OK")
	}

	// Test CreateVolume
	rand.Seed(time.Now().Unix())
	volumename := fmt.Sprintf("csivolume-%v", rand.Int())
	volumeinfo, err := controller.CreateVolume(context.Background(),
		versions[0], volumename, nil, nil, nil, nil)
	if err != nil {
		log.Fatalf("failed to CreateVolume: %v", err)
	} else {
		log.Printf("[CreateVolume] CreateVolume:%v", volumeinfo)
	}

	// Test ControllerPublishVolume
	publishvolumeinfo, err := controller.ControllerPublishVolume(context.Background(),
		versions[0], volumeinfo.Id, nodeid, nil, false, nil, volumeinfo.Attributes)
	if err != nil {
		log.Fatalf("failed to ControllerPublishVolume: %v", err)
	} else {
		log.Printf("[ControllerPublishVolume] ControllerPublishVolume:%v", publishvolumeinfo)
	}

	// Test NodePublishVolume
	targetpath := "/var/lib/kubelet/plugins/opensds/"
	err = node.NodePublishVolume(context.Background(),
		versions[0], volumeinfo.Id, publishvolumeinfo,
		targetpath, nil, false, nil, nil)
	if err != nil {
		log.Fatalf("failed to NodePublishVolume: %v", err)
	} else {
		log.Printf("[NodePublishVolume] NodePublishVolume:OK")
	}

	// Set SleepTime for checking the status of Volume
	sleeptime := os.Getenv("SLEEPTIME")
	nsleeptime := 1
	if len(sleeptime) > 0 {
		nsleeptime, _ = strconv.Atoi(sleeptime)
	}
	log.Printf("[SleepTime] %v Seconds", nsleeptime)
	time.Sleep(time.Duration(nsleeptime) * time.Second)

	// Test NodeUnpublishVolume
	err = node.NodeUnpublishVolume(context.Background(),
		versions[0], volumeinfo.Id, targetpath, nil)
	if err != nil {
		log.Fatalf("failed to NodeUnpublishVolume: %v", err)
	} else {
		log.Printf("[NodeUnpublishVolume] NodeUnpublishVolume:OK")
	}

	// Test ControllerUnpublishVolume
	err = controller.ControllerUnpublishVolume(context.Background(), versions[0], volumeinfo.Id, nodeid, nil)
	if err != nil {
		log.Fatalf("failed to ControllerUnpublishVolume: %v", err)
	} else {
		log.Printf("[ControllerUnpublishVolume] ControllerUnpublishVolume:OK")
	}

	// Test DeleteVolume
	err = controller.DeleteVolume(context.Background(), versions[0], volumeinfo.Id, nil)
	if err != nil {
		log.Fatalf("failed to DeleteVolume: %v", err)
	} else {
		log.Printf("[DeleteVolume] DeleteVolume:%v", volumeinfo)
	}
}
