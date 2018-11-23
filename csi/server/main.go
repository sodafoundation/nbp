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
	"flag"
	"os"
	"os/signal"
	"syscall"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/opensds/nbp/csi/server/plugin"
	"github.com/opensds/nbp/csi/server/plugin/opensds"
	"github.com/opensds/nbp/csi/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/opensds/opensds/contrib/connector/fc"
	_ "github.com/opensds/opensds/contrib/connector/iscsi"
	_ "github.com/opensds/opensds/contrib/connector/rbd"
)

var (
	csiEndpoint         string
	opensdsEndpoint     string
	opensdsAuthStrategy string
)

func init() {
	flag.Set("alsologtostderr", "true")
}

func main() {

	flag.CommandLine.Parse([]string{})
	// Open OpenSDS dock service log file.
	util.InitLogs()
	defer util.FlushLogs()

	cmd := &cobra.Command{
		Use:   "OpenSDS",
		Short: "CSI based OpenSDS driver",
		Run: func(cmd *cobra.Command, args []string) {
			handle()
		},
	}

	// the endpoint variable priority is flag, ENV and default.
	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	csiEp := util.CSIDefaultEndpoint
	opensdsEp := util.OpensdsDefaultEndpoint
	if ep, ok := os.LookupEnv(util.CSIEndpoint); ok {
		csiEp = ep
	}
	if ep, ok := os.LookupEnv(util.OpensdsEndpoint); ok {
		opensdsEp = ep
	}
	cmd.PersistentFlags().StringVar(&csiEndpoint, "csiEndpoint", csiEp, "CSI Endpoint")
	cmd.PersistentFlags().StringVar(&opensdsEndpoint, "opensdsEndpoint", opensdsEp, "OpenSDS Endpoint")
	cmd.PersistentFlags().StringVar(&opensdsAuthStrategy, "opensdsAuthStrategy", "", "OpenSDS Auth Strategy")

	cmd.ParseFlags(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		glog.Errorf("failed to execute: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func handle() {

	// Set Env
	os.Setenv(util.OpensdsEndpoint, opensdsEndpoint)
	os.Setenv(util.OpensdsAuthStrategy, opensdsAuthStrategy)

	// Get CSI Endpoint Listener
	lis, err := util.GetCSIEndPointListener(csiEndpoint)
	if err != nil {
		glog.Errorf("failed to listen: %v", err)
	}

	// New Grpc Server
	s := grpc.NewServer()

	// Register CSI Service
	var defaultplugin plugin.Service = &opensds.Plugin{}
	conServer := &server{plugin: defaultplugin}
	csi.RegisterIdentityServer(s, conServer)
	csi.RegisterControllerServer(s, conServer)
	csi.RegisterNodeServer(s, conServer)

	// Register reflection Service
	reflection.Register(s)

	// Remove sock file
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			if sig == syscall.SIGKILL ||
				sig == syscall.SIGQUIT ||
				sig == syscall.SIGHUP ||
				sig == syscall.SIGTERM ||
				sig == syscall.SIGINT {
				glog.Info("exit to serve")
				if lis.Addr().Network() == "unix" {
					sockfile := lis.Addr().String()
					os.RemoveAll(sockfile)
					glog.Infof("remove sock file: %s", sockfile)
				}
				os.Exit(0)
			}
		}
	}()

	// Serve Plugin Server
	glog.Infof("start to serve: %s", lis.Addr())
	if err := s.Serve(lis); err != nil {
		glog.Errorf("failed to serve: %v", err)
	}
}
