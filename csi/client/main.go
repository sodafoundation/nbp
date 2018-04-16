package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/opensds/nbp/csi/client/proxy"
	"github.com/spf13/cobra"
)

var (
	csiEndpoint string
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {

	flag.CommandLine.Parse([]string{})

	cmd := &cobra.Command{
		Use:   "OpenSDS",
		Short: "CSI based OpenSDS driver client",
		Run: func(cmd *cobra.Command, args []string) {
			handle()
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(&csiEndpoint, "csiEndpoint", "", "CSI Endpoint")

	cmd.ParseFlags(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		log.Fatalf("failed to execute: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func handle() {

	// Set Env
	os.Setenv("CSI_ENDPOINT", csiEndpoint)

	// GetIdentity
	log.Println("start to get identity")
	identity, err := proxy.GetIdentity()
	if err != nil {
		log.Fatalf("failed to get identity: %v", err)
	}

	// Test GetPluginInfo
	rs, err := identity.GetPluginInfo(context.Background())
	if err != nil {
		log.Fatalf("failed to GetPluginInfo: %v", err)
	}

	log.Printf("[GetPluginInfoResponse] Name:%s VendorVersion:%s Manifest:%v",
		rs.Name, rs.VendorVersion, rs.Manifest)

	// Test GetPluginCapabilities
	rs1, err := identity.GetPluginCapabilities(context.Background())
	if err != nil {
		log.Fatalf("failed to GetPluginCapabilities: %v", err)
	}

	log.Printf("[GetPluginCapabilities] capabilites:%v", rs1)

	// Test Probe
	rs2, err := identity.Probe(context.Background())
	if err != nil {
		log.Fatalf("failed to Probe: %v", err)
	}

	log.Printf("[Probe] response:%v", rs2)

	// GetController
	log.Println("start to get controller")
	controller, err := proxy.GetController()
	if err != nil {
		log.Fatalf("failed to get controller: %v", err)
	}

	// Test ControllerGetCapabilities
	controllercapabilities, err := controller.ControllerGetCapabilities(context.Background())
	if err != nil {
		log.Fatalf("failed to ControllerGetCapabilities: %v", err)
	}

	log.Printf("[ControllerGetCapabilities] controllercapabilities:%v", controllercapabilities)

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
	nodecapabilities, err := node.NodeGetCapabilities(context.Background())
	if err != nil {
		log.Fatalf("failed to NodeGetCapabilities: %v", err)
	}

	log.Printf("[NodeGetCapabilities] nodecapabilities:%v", nodecapabilities)

	// Test GetNodeID
	nodeid, err := node.NodeGetId(context.Background())
	if err != nil {
		log.Fatalf("failed to NodeGetId: %v", err)
	}

	log.Printf("[NodeGetId] nodeid:%v", nodeid)

	// Test CreateVolume
	rand.Seed(time.Now().Unix())
	volumename := fmt.Sprintf("csivolume-%v", rand.Int())
	volumeinfo, err := controller.CreateVolume(context.Background(), volumename, nil, nil, nil, nil)
	if err != nil {
		log.Fatalf("failed to CreateVolume: %v", err)
	} else {
		log.Printf("[CreateVolume] CreateVolume:%v", volumeinfo)
	}

	// Test ControllerPublishVolume
	publishvolumeinfo, err := controller.ControllerPublishVolume(context.Background(),
		volumeinfo.Id, nodeid, nil, false, nil, volumeinfo.Attributes)
	if err != nil {
		log.Fatalf("failed to ControllerPublishVolume: %v", err)
	} else {
		log.Printf("[ControllerPublishVolume] ControllerPublishVolume:%v", publishvolumeinfo)
	}

	// Test NodePublishVolume
	targetpath := "/var/lib/kubelet/plugins/opensds/"
	err = node.NodePublishVolume(context.Background(),
		volumeinfo.Id, publishvolumeinfo, "",
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
		volumeinfo.Id, targetpath)
	if err != nil {
		log.Fatalf("failed to NodeUnpublishVolume: %v", err)
	} else {
		log.Printf("[NodeUnpublishVolume] NodeUnpublishVolume:OK")
	}

	// Test ControllerUnpublishVolume
	err = controller.ControllerUnpublishVolume(context.Background(), volumeinfo.Id, nodeid, nil)
	if err != nil {
		log.Fatalf("failed to ControllerUnpublishVolume: %v", err)
	} else {
		log.Printf("[ControllerUnpublishVolume] ControllerUnpublishVolume:OK")
	}

	// Test DeleteVolume
	err = controller.DeleteVolume(context.Background(), volumeinfo.Id, nil)
	if err != nil {
		log.Fatalf("failed to DeleteVolume: %v", err)
	} else {
		log.Printf("[DeleteVolume] DeleteVolume:%v", volumeinfo)
	}
}
