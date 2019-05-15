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

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	plugin "github.com/opensds/nbp/csi/server/plugin/opensds"
	"github.com/opensds/nbp/csi/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/opensds/opensds/contrib/connector/fc"
	_ "github.com/opensds/opensds/contrib/connector/iscsi"
	_ "github.com/opensds/opensds/contrib/connector/rbd"
)

func main() {
	flag.Set("alsologtostderr", "true")
	// Open OpenSDS dock service log file.
	util.InitLogs()
	defer util.FlushLogs()

	var csiEndpoint, opensdsEndpoint, opensdsAuthStrategy string
	// CSI endpoint
	flag.StringVar(&csiEndpoint, "csiEndpoint", util.CSIDefaultEndpoint, "CSI Endpoint")

	// opensds endpoint
	flag.StringVar(&opensdsEndpoint, "opensdsEndpoint", util.OpensdsDefaultEndpoint, "OpenSDS Endpoint")

	// opensds auth strategy
	flag.StringVar(&opensdsAuthStrategy, "opensdsAuthStrategy", util.OpensdsDefaultAuthStrategy, "OpenSDS Auth Strategy")

	flag.Parse()

	// Get CSI Endpoint Listener
	lis, err := util.GetCSIEndPointListener(csiEndpoint)
	if err != nil {
		glog.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}

	// Initialize the driver
	pluginServer, err := plugin.NewServer(opensdsEndpoint, opensdsAuthStrategy)
	if err != nil {
		glog.Errorf("failed to initialize the driver: %v", err)
		os.Exit(1)
	}

	// New Grpc Server
	s := grpc.NewServer()

	// Register CSI Service
	csi.RegisterIdentityServer(s, pluginServer)
	csi.RegisterControllerServer(s, pluginServer)
	csi.RegisterNodeServer(s, pluginServer)

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
