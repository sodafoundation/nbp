// Copyright 2018 The OpenSDS Authors.
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
	"os"

	"github.com/golang/glog"
	_ "github.com/sodafoundation/dock/contrib/connector/fc"
	_ "github.com/sodafoundation/dock/contrib/connector/iscsi"
	_ "github.com/sodafoundation/dock/contrib/connector/nfs"
	_ "github.com/sodafoundation/dock/contrib/connector/nvmeof"
	_ "github.com/sodafoundation/dock/contrib/connector/rbd"
	"github.com/sodafoundation/nbp/csi/common"
	"github.com/sodafoundation/nbp/csi/plugins/file"
)

func main() {
	// Initialise plugin parameters.
	client, listener, err := common.InitPlugin()
	if err != nil {
		glog.Errorf("failed to initialise csi file plugin: %v", err)
		os.Exit(1)
	}

	// Initialize the driver
	pluginServer, err := file.NewServer(client)
	if err != nil {
		glog.Errorf("failed to initialize the driver: %v", err)
		os.Exit(1)
	}

	// Start grpc server to listen and handle CSI invocations
	common.NewGrpcServer(listener, pluginServer)
}
