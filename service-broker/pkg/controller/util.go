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

	"github.com/golang/glog"
	sdsClient "github.com/opensds/opensds/client"
	ctx "github.com/opensds/opensds/pkg/context"
	dockClient "github.com/opensds/opensds/pkg/dock/client"
	"github.com/opensds/opensds/pkg/model"
	pb "github.com/opensds/opensds/pkg/model/proto"
	"golang.org/x/net/context"
)

const (
	attacherDockFlag = "attacher"
)

type DeviceSpec map[string]string

func AttachVolume(c *sdsClient.Client, in *model.VolumeAttachmentSpec) (DeviceSpec, error) {
	dck, err := discoverAttacherDock(c, in.HostInfo.Host)
	if err != nil {
		glog.Errorf("failed to find attach dock with nodeID(%s): %v", in.HostInfo.Host, err)
		return nil, err
	}
	// Create attach dock client and connect attach dock server.
	dckClient := dockClient.NewClient()
	if err = dckClient.Connect(dck.Endpoint); err != nil {
		glog.Errorf("failed to connect attach dock with endpoint(%s): %v", dck.Endpoint, err)
		return nil, err
	}
	defer dckClient.Close()

	connData, _ := json.Marshal(in.ConnectionData)
	var attachOpt = &pb.AttachVolumeOpts{
		AccessProtocol: in.DriverVolumeType,
		ConnectionData: string(connData),
		Metadata:       map[string]string{},
		Context:        ctx.NewAdminContext().ToJson(),
	}
	response, err := dckClient.AttachVolume(context.Background(), attachOpt)
	if err != nil {
		glog.Errorf("failed to attach dock volume to host: %v", err)
		return nil, err
	}

	return DeviceSpec{"device": response.GetResult().GetMessage()}, nil
}

func DetachVolume(c *sdsClient.Client, in *model.VolumeAttachmentSpec) error {
	dck, err := discoverAttacherDock(c, in.HostInfo.Host)
	if err != nil {
		glog.Errorf("failed to find attach dock with nodeID(%s): %v", in.HostInfo.Host, err)
		return err
	}
	// Create attach dock client and connect attach dock server.
	dckClient := dockClient.NewClient()
	if err = dckClient.Connect(dck.Endpoint); err != nil {
		glog.Errorf("failed to connect attach dock with endpoint(%s): %v", dck.Endpoint, err)
		return err
	}
	defer dckClient.Close()

	connData, _ := json.Marshal(in.ConnectionData)
	var detachOpt = &pb.DetachVolumeOpts{
		AccessProtocol: in.DriverVolumeType,
		ConnectionData: string(connData),
		Metadata:       map[string]string{},
		Context:        ctx.NewAdminContext().ToJson(),
	}

	_, err = dckClient.DetachVolume(context.Background(), detachOpt)
	return err
}

func discoverAttacherDock(c *sdsClient.Client, nodeId string) (*model.DockSpec, error) {
	dcks, err := c.ListDocks()
	if err != nil {
		glog.Errorf("failed to list docks: %v", err)
		return nil, err
	}

	d := func() *model.DockSpec {
		for _, dck := range dcks {
			if dck.Type == attacherDockFlag && dck.NodeId == nodeId {
				return dck
			}
		}
		return nil
	}()
	if d == nil {
		return nil, fmt.Errorf("Can't find supported attach dock (%s)", nodeId)
	}

	return d, nil
}
