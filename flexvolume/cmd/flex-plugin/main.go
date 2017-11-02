// Copyright (c) 2016 Huawei Technologies Co., Ltd. All Rights Reserved.
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

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/opensds/nbp/flexvolume/pkg/api"
	"github.com/opensds/nbp/flexvolume/pkg/connector"
	"github.com/opensds/nbp/flexvolume/pkg/volume"
)

type OpenSDSOptions struct {
	DefaultOptions
	VolumeId      string `json:"volumeId"`
	BackendDriver string `json:"backendDriver"`
}

type OpenSDSPlugin struct{}

func (OpenSDSPlugin) Init() Result {
	return Succeed()
}

func (OpenSDSPlugin) NewOptions() interface{} {
	var option = &OpenSDSOptions{}
	return option
}

func (OpenSDSPlugin) Attach(opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)
	volID := opt.VolumeId

	prop, err := connector.GetConnectorProperties(false)
	if err != nil {
		return Fail(err.Error())
	}

	atc, err := volume.CreateVolumeAttachment(volID, prop)
	if err != nil {
		return Fail(err.Error())
	}
	conn := &connector.Connector{
		ConnInfo: atc.ConnectionInfo,
	}

	// log.Printf("Receive connection info: %+v\n", conn)
	devPath, err := conn.ConnectVolume()
	if err != nil {
		return Fail(err.Error())
	}

	_, err = volume.UpdateVolumeAttachment(atc.Id, volID, devPath, atc.HostInfo)
	if err != nil {
		return Fail(err.Error())
	} else {
		return Result{
			Status: "Success",
			Device: devPath,
		}
	}
}

func (OpenSDSPlugin) Detach(device string) Result {
	linkPath, err := FindLinkPath(device)
	if err != nil {
		return Fail(err.Error())
	}
	ind := strings.Index(linkPath, "by-id/")
	if ind <= 0 {
		return Fail(fmt.Errorf("Detach disk: no volume id in %s", linkPath))
	}

	var volID = linkPath[ind+6:]

	isMultipath := false
	prop, err := connector.GetConnectorProperties(isMultipath)
	if err != nil {
		return Fail(err.Error())
	}

	atcs, err := volume.ListVolumeAttachments(volID)
	if err != nil {
		return Fail(err.Error())
	}
	atcFound, atcPtr := false, &api.VolumeAttachment{}
	for _, atc := range *atcs {
		if atc.Mountpoint == linkPath && atc.HostInfo.Host == prop.Host {
			atcFound, atcPtr = true, &atc
		}
	}
	if !atcFound {
		return Fail("Wrong device path, can not find volume attachment!")
	}

	conn := &connector.Connector{
		ConnInfo: atcPtr.ConnectionInfo,
	}
	_, err = conn.DisconnectVolume()
	if err != nil {
		return Fail(err.Error())
	}

	volumeResponse, err := volume.DeleteVolumeAttachment(atcPtr.Id, volID)
	if err != nil {
		return Fail(err.Error())
	} else {
		if volumeResponse.Status == "Success" {
			return Succeed()
		} else {
			err = errors.New("Detach volume failed!")
			return Fail(err.Error())
		}
	}
}

func (OpenSDSPlugin) Mount(mountDir string, device string, opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)

	_, err := volume.MountVolume(mountDir, device, opt.FsType)
	if err != nil {
		return Fail(err.Error())
	}
	return Succeed()
}

func (OpenSDSPlugin) Unmount(mountDir string) Result {
	_, err := volume.UnmountVolume(mountDir)
	if err != nil {
		return Fail(err.Error())
	}
	return Succeed()
}

func main() {
	RunPlugin(&OpenSDSPlugin{})
}

