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

/*
This module implements utility functionalities to all OpenSDS northbound service.
*/

package util

import (
	"errors"
	"fmt"
	"github.com/sodafoundation/api/client"
	"github.com/sodafoundation/api/pkg/model"
)

// GetHostByHostName returns hostid for given hostname if found
func GetHostByHostName(client *client.Client, hostName string) (*model.HostSpec, error) {
	hostLists, err := client.HostMgr.ListHosts()
	if nil != err {
		return nil, errors.New("Failed to list hosts")
	}

	for _, elem := range hostLists {
		if elem.HostName == hostName {
			return elem, nil
		}
	}
	return nil, fmt.Errorf("Failed to find host for hostname %s", hostName)
}

// GetHostByHostId returns hostid for given hostId if found
func GetHostByHostId(client *client.Client, hostId string) (*model.HostSpec, error) {
	hostLists, err := client.HostMgr.ListHosts()
	if nil != err {
		return nil, errors.New("Failed to list hosts")
	}

	for _, elem := range hostLists {
		if elem.Id == hostId {
			return elem, nil
		}
	}
	return nil, fmt.Errorf("Failed to find host for hostId %s", hostId)
}
