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
	"github.com/golang/glog"
	opensdsClient "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/csi/server/plugin"
	"github.com/opensds/opensds/client"
)

const (
	// PluginName setting
	PluginName      = "csi-opensdsplugin"
	FakeIQN         = "fakeIqn"
	TopologyZoneKey = "topology." + PluginName + "/zone"
)

// Plugin define
type Plugin struct {
	Client *client.Client
}

func NewServer(endpoint, authStrategy string) (plugin.Service, error) {
	// get opensds client
	client, err := opensdsClient.GetClient(endpoint, authStrategy)
	if client == nil || err != nil {
		glog.Errorf("get opensds client failed: %v", err)
		return nil, err
	}

	p := &Plugin{Client: client}

	// When there are multiple volumes unmount at the same time,
	// it will cause conflicts related to the state machine,
	// so start a watch list to let the volumes unmount one by one.
	go p.UnpublishRoutine()

	return p, nil
}
