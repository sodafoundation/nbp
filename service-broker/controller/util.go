// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package controller

import (
	"encoding/json"
	"fmt"

	sdsClient "github.com/opensds/opensds/client"
	c "github.com/opensds/opensds/pkg/context"
	dockClient "github.com/opensds/opensds/pkg/dock/client"
	pb "github.com/opensds/opensds/pkg/dock/proto"
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
)

type DeviceSpec map[string]string

func AttachVolume(edp string, in *model.VolumeAttachmentSpec) (DeviceSpec, error) {
	dckClient, err := discoverAttachDock(edp, in.HostInfo.Host)
	if err != nil {
		return nil, err
	}
	defer dckClient.Close()

	connData, _ := json.Marshal(in.ConnectionData)
	var attachOpt = &pb.AttachVolumeOpts{
		AccessProtocol: in.DriverVolumeType,
		ConnectionData: string(connData),
		Metadata:       map[string]string{},
		Context:        c.NewAdminContext().ToJson(),
	}
	response, err := dckClient.AttachVolume(context.Background(), attachOpt)
	if err != nil {
		return nil, err
	}

	return DeviceSpec{"device": response.GetResult().GetMessage()}, nil
}

func DetachVolume(edp string, in *model.VolumeAttachmentSpec) error {
	dckClient, err := discoverAttachDock(edp, in.HostInfo.Host)
	if err != nil {
		return err
	}
	defer dckClient.Close()

	connData, _ := json.Marshal(in.ConnectionData)
	var detachOpt = &pb.DetachVolumeOpts{
		AccessProtocol: in.DriverVolumeType,
		ConnectionData: string(connData),
		Metadata:       map[string]string{},
		Context:        c.NewAdminContext().ToJson(),
	}
	response, err := dckClient.DetachVolume(context.Background(), detachOpt)
	if err != nil {
		return err
	}

	return nil
}

func discoverAttachDock(edp, nodeId string) (dockClient.Client, error) {
	dcks, err := sdsClient.NewClient(&sdsClient.Config{Endpoint: edp}).
		ListDocks()
	if err != nil {
		return nil, err
	}

	d := func() *model.DockSpec {
		for _, dck := range dcks {
			if dck.NodeId == nodeId {
				return dck
			}
		}
		return nil
	}()
	if d == nil {
		return nil, fmt.Errorf("Can't find supported attach dock (%s)", nodeId)
	}
	// Create attach dock client and connect attach dock server.
	dckClient := dockClient.NewClient()
	if dckClient.Connect(d.Endpoint); err != nil {
		return nil, err
	}
	return dckClient, nil
}

func ConvertToHostInfoStruct(mapObj interface{}) (*model.HostInfo, error) {
	jsonStr, err := json.Marshal(mapObj)
	if nil != err {
		return nil, err
	}

	var result *model.HostInfo
	if err = json.Unmarshal(jsonStr, result); err != nil {
		return nil, err
	}

	return result, nil
}
