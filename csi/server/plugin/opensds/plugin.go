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

package opensds

import (
	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	opensdsClient "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/opensds/client"
)

const (
	// PluginName setting
	PluginName = "csi-opensdsplugin"
	FakeIQN    = "fakeIqn"
)

// Plugin define
type Plugin struct {
	Cli *client.Client
}

// Service Define CSI Interface
type Service interface {
	csi.IdentityServer
	csi.ControllerServer
	csi.NodeServer
}

func NewServer() (Service, error) {
	client, err := opensdsClient.GetClient("", "")
	if client == nil || err != nil {
		glog.Errorf("get opensds client failed: %v", err)
		return nil, err
	}

	return &Plugin{Cli: client}, nil
}
