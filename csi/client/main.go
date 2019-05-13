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

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/opensds/nbp/csi/client/proxy"
	"github.com/opensds/nbp/csi/server/plugin/opensds"
	"github.com/opensds/nbp/csi/util"
	"github.com/spf13/cobra"
)

var (
	csiEndpoint       string
	enableReplication bool
)

var (
	node       proxy.Node
	controller proxy.Controller
)

const targetpath = "/var/lib/kubelet/plugins/opensds/"

func init() {
	flag.Set("alsologtostderr", "true")
}

func main() {
	flag.CommandLine.Parse([]string{})

	// Open OpenSDS dock service log file.
	util.InitLogs()
	defer util.FlushLogs()

	rootCmd := &cobra.Command{
		Use:   "OpenSDS",
		Short: "CSI based OpenSDS driver client",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
			os.Exit(1)
		},
	}
	constructSubCmd(rootCmd)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// the endpoint variable priority is flag, ENV and default.
	csiEp := util.CSIDefaultEndpoint
	if ep, ok := os.LookupEnv(util.CSIEndpoint); ok {
		csiEp = ep
	}
	rootCmd.PersistentFlags().StringVar(&csiEndpoint, "csiEndpoint", csiEp, "CSI Endpoint")

	// EnableReplication
	enableReplication = false
	if v, ok := os.LookupEnv(util.CSIEnableReplication); ok {
		if strings.ToLower(v) == "true" {
			enableReplication = true
		}
	}
	rootCmd.PersistentFlags().BoolVar(&enableReplication, "enableReplication",
		enableReplication, "Weather enable the replication")

	rootCmd.ParseFlags(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		glog.Errorf("failed to execute: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

var allCommand = &cobra.Command{
	Use:   "all",
	Short: "test full process",
	Run:   allAction,
}

var queryCommand = &cobra.Command{
	Use:   "query",
	Short: "query test",
	Run:   queryTestAction,
}

var volumeCommand = &cobra.Command{
	Use:   "volume",
	Short: "volume test",
	Run:   volumeTestAction,
}

var controllerPublishCommand = &cobra.Command{
	Use:   "controller",
	Short: "controller publish test",
	Run:   controllerPublishTestAction,
}

var nodePublishCommand = &cobra.Command{
	Use:   "node",
	Short: "node publish test",
	Run:   nodePublishTestAction,
}

var stagePublishCommand = &cobra.Command{
	Use:   "stage",
	Short: "get current stage",
	Run:   getStageAction,
}

var (
	cleanFlag bool
)

func constructSubCmd(rootCmd *cobra.Command) {
	volumeCommand.Flags().BoolVarP(&cleanFlag, "cleanup", "c", false, "cleanup this resrouce")
	controllerPublishCommand.Flags().BoolVarP(&cleanFlag, "cleanup", "c", false, "cleanup this resrouce")
	nodePublishCommand.Flags().BoolVarP(&cleanFlag, "cleanup", "c", false, "cleanup this resrouce")
	// add sub command
	rootCmd.AddCommand(allCommand)
	rootCmd.AddCommand(queryCommand)
	rootCmd.AddCommand(volumeCommand)
	rootCmd.AddCommand(controllerPublishCommand)
	rootCmd.AddCommand(nodePublishCommand)
	rootCmd.AddCommand(stagePublishCommand)
}

func allAction(cmd *cobra.Command, args []string) {
	setup()
	allHandler()
}

func queryTestAction(cmd *cobra.Command, args []string) {
	setup()
	queryCheck()
}

type action func()

func stageWraper(srcStage, destStage string, action action) {

	name2enum := map[string]int{
		stageStart:             0,
		stageVolume:            1,
		stageControllerPublish: 2,
		stageNodePublish:       3,
	}

	curStage, _ := GetTestStage()
	if srcStage != curStage {
		glog.Errorf("Action got a wrong stage direction, %s ==> %s", curStage, destStage)
		os.Exit(-1)
	}

	curStageN := name2enum[curStage]
	nextStageN := name2enum[destStage]
	diff := curStageN - nextStageN
	if diff != 1 && diff != -1 {
		glog.Errorf("Action got a wrong stage direction, %s ==> %s", curStage, destStage)
		os.Exit(-1)
	}

	setup()
	action()
	SetTestStage(destStage)
}

func volumeTestAction(cmd *cobra.Command, args []string) {
	if !cleanFlag {
		stageWraper(stageStart, stageVolume, volumeCreate)
	} else {
		stageWraper(stageVolume, stageStart, volumeDelete)
	}
}

func controllerPublishTestAction(cmd *cobra.Command, args []string) {
	if !cleanFlag {
		stageWraper(stageVolume, stageControllerPublish, controllerPublishVolume)
	} else {
		stageWraper(stageControllerPublish, stageVolume, controllerUnpublishVolume)
	}
}

func nodePublishTestAction(cmd *cobra.Command, args []string) {
	if !cleanFlag {
		stageWraper(stageControllerPublish, stageNodePublish, nodePublishVolume)
	} else {
		stageWraper(stageNodePublish, stageControllerPublish, nodeUnpulishVolume)
	}
}

func getStageAction(cmd *cobra.Command, args []string) {
	stage, err := GetTestStage()
	if err != nil {
		glog.Error(err)
	}
	glog.Info("Current Stage: ", stage)
}

func setup() {
	// GetController
	var err error
	controller, err = proxy.GetController(csiEndpoint)
	if err != nil {
		glog.Fatalf("failed to get controller: %v", err)
	}

	// GetNode
	node, err = proxy.GetNode(csiEndpoint)
	if err != nil {
		glog.Fatalf("failed to get node: %v", err)
	}
}

func queryCheck() {

	// GetIdentity
	glog.Info("start to get identity")
	identity, err := proxy.GetIdentity(csiEndpoint)
	if err != nil {
		glog.Fatalf("failed to get identity: %v", err)
	}

	// Test GetPluginInfo
	rs, err := identity.GetPluginInfo(context.Background())
	if err != nil {
		glog.Fatalf("failed to GetPluginInfo: %v", err)
	}

	glog.Infof("[GetPluginInfoResponse] Name:%s VendorVersion:%s Manifest:%v",
		rs.Name, rs.VendorVersion, rs.Manifest)

	// Test GetPluginCapabilities
	rs1, err := identity.GetPluginCapabilities(context.Background())
	if err != nil {
		glog.Fatalf("failed to GetPluginCapabilities: %v", err)
	}

	glog.Infof("[GetPluginCapabilities] capabilites:%v", rs1)

	// Test Probe
	rs2, err := identity.Probe(context.Background())
	if err != nil {
		glog.Fatalf("failed to Probe: %v", err)
	}

	glog.Infof("[Probe] response:%v", rs2)

	// Test ControllerGetCapabilities
	controllercapabilities, err := controller.ControllerGetCapabilities(context.Background())
	if err != nil {
		glog.Fatalf("failed to ControllerGetCapabilities: %v", err)
	}

	glog.Infof("[ControllerGetCapabilities] controllercapabilities:%v", controllercapabilities)

	// Test ValidateVolumeCapabilities
	volumeid := "1234567890"
	volumecapabilities := []*csi.VolumeCapability{
		{
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
		glog.Fatalf("failed to ValidateVolumeCapabilities: %v", err)
	}

	glog.Infof("[ValidateVolumeCapabilities] validatevolumecapabilities:%v", validatevolumecapabilities)

	// Test NodeGetCapabilities
	nodecapabilities, err := node.NodeGetCapabilities(context.Background())
	if err != nil {
		glog.Fatalf("failed to NodeGetCapabilities: %v", err)
	}

	glog.Infof("[NodeGetCapabilities] nodecapabilities:%v", nodecapabilities)
}

func volumeCreate() {
	// Test CreateVolume
	rand.Seed(time.Now().Unix())
	volumename := fmt.Sprintf("csivolume-%v", rand.Int())

	param := map[string]string{opensds.KParamSecondaryAZ: util.OpensdsDefaultSecondaryAZ}
	// add to param map if the replication is set.
	if enableReplication {
		param[opensds.KParamEnableReplication] = "true"
	}
	// set the secondary az
	if v, ok := os.LookupEnv(util.OpensdsSecondaryAZ); ok {
		param[opensds.KParamSecondaryAZ] = v
	}
	volumeinfo, err := controller.CreateVolume(context.Background(), volumename, nil, nil,
		param, nil)
	if err != nil {
		glog.Fatalf("failed to CreateVolume: %v", err)
	} else {
		glog.Infof("[CreateVolume] CreateVolume:%v", volumeinfo)
	}
	if err := StoreVolume(volumeinfo); err != nil {
		glog.Error("Store volume info in db failed,", err)
	}
	glog.Info("CreateVolume: OK")
}

func volumeDelete() {
	// Test DeleteVolume
	volumeinfo, err := GetVolume()
	if err != nil {
		glog.Fatalf("Get volume info failed: %v", err)
	}
	err = controller.DeleteVolume(context.Background(), volumeinfo.VolumeId, nil)
	if err != nil {
		glog.Fatalf("failed to DeleteVolume: %v", err)
	} else {
		glog.Infof("[DeleteVolume] DeleteVolume:%v", volumeinfo)
		glog.Info("DeleteVolume: OK")
	}
}

func controllerPublishVolume() {
	// Test NodeGetInfo
	nodeid, err := node.NodeGetInfo(context.Background())
	if err != nil {
		glog.Fatalf("failed to call NodeGetInfo: %v", err)
	}
	glog.Infof("[NodeGetInfo] nodeid:%v", nodeid)

	volumeinfo, err := GetVolume()
	if err != nil {
		glog.Fatalf("Get volume info failed: %v", err)
	}

	// Test ControllerPublishVolume
	publishvolumeinfo, err := controller.ControllerPublishVolume(context.Background(),
		volumeinfo.VolumeId, nodeid, nil, false, nil, volumeinfo.VolumeContext)
	if err != nil {
		glog.Fatalf("failed to ControllerPublishVolume: %v", err)
	} else {
		glog.Infof("[ControllerPublishVolume] ControllerPublishVolume:%v", publishvolumeinfo)
	}

	if err := StorePublishVolume(publishvolumeinfo); err != nil {
		glog.Error("Store publish volume info in db failed,", err)
	} else {
		glog.Info("ControllerPublishVolume: OK")
	}
}

func controllerUnpublishVolume() {
	// Test GetNodeID
	volumeinfo, err := GetVolume()
	if err != nil {
		glog.Fatalf("Get volume info failed: %v", err)
	}
	nodeid, err := node.NodeGetInfo(context.Background())
	if err != nil {
		glog.Fatalf("failed to call NodeGetInfo: %v", err)
	}
	glog.Infof("[NodeGetInfo] nodeid:%v", nodeid)
	// Test ControllerUnpublishVolume
	err = controller.ControllerUnpublishVolume(context.Background(), volumeinfo.VolumeId, nodeid, nil)
	if err != nil {
		glog.Fatalf("failed to ControllerUnpublishVolume: %v", err)
	} else {
		glog.Infof("ControllerUnpublishVolume: OK")
	}

}

func nodePublishVolume() {
	// Test NodePublishVolume
	volumeinfo, err := GetVolume()
	if err != nil {
		glog.Fatalf("Get volume info failed: %v", err)
	}
	publishvolumeinfo, err := GetPublishVolume()
	if err != nil {
		glog.Fatalf("Get publish volume info failed %v", err)
	}
	err = node.NodePublishVolume(context.Background(),
		volumeinfo.VolumeId, publishvolumeinfo, "",
		targetpath, nil, false, nil, nil)
	if err != nil {
		glog.Fatalf("failed to NodePublishVolume: %v", err)
	} else {
		glog.Infof("NodePublishVolume:OK")
	}
}

func nodeUnpulishVolume() {
	// Test NodeUnpublishVolume
	volumeinfo, err := GetVolume()
	if err != nil {
		glog.Fatalf("Get volume info failed %v", err)
	}
	err = node.NodeUnpublishVolume(context.Background(),
		volumeinfo.VolumeId, targetpath)
	if err != nil {
		glog.Fatalf("failed to NodeUnpublishVolume: %v", err)
	} else {
		glog.Infof("NodeUnpublishVolume:OK")
	}
}

func allHandler() {
	defer glog.Info("Testing end")
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Println("Press any key to start test ...")
	inputReader.ReadByte()

	queryCheck()
	volumeCreate()
	controllerPublishVolume()
	nodePublishVolume()

	// Suspend util user enter a key.
	fmt.Println("Press any key to continue...")
	inputReader.ReadByte()

	nodeUnpulishVolume()
	controllerUnpublishVolume()
	volumeDelete()

	RemoveDbFile()
}
