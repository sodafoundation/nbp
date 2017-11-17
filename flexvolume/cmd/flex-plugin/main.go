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

	"github.com/opensds/nbp/client/iscsi"
	"github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/flexvolume/pkg/volume"
	"github.com/opensds/opensds/pkg/model"
	"os"
	"runtime"
	"strconv"
)

//TODO: if volume has status, opensds should supply the definition of status, and set status of volume.
var (
	VOLUME_STATUS_ATTACHING = "attaching"
	VOLUME_STATUS_ATTACHED  = "attached"
	VOLUME_STATUS_DETACHING = "detaching"
	VOLUME_STATUS_DETACHED  = "detached"
)

type OpenSDSOptions struct {
	VolumeId   string `json:"kubernetes.io/pvOrVolumeName"`
	AccessMode string `json:"kubernetes.io/readwrite`
	FsType     string `json:"kubernetes.io/fsType"`
}

type OpenSDSPlugin struct{}

func (plugin *OpenSDSPlugin) Init() Result {
	return Succeed()
}

func (plugin *OpenSDSPlugin) NewOptions() interface{} {
	var option = &OpenSDSOptions{}
	return option
}

func (plugin *OpenSDSPlugin) Attach(opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)
	volID := opt.VolumeId

	client := opensds.GetClient("")
	_, errVol := client.GetVolume(volID)
	if errVol != nil {
		return Fail(errors.New("volume not exist."))
	}

	hostname, _ := os.Hostname()
	//get attachment to check whether the volume can attach to this node.
	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return Fail(err.Error())
	}

	for _, value := range attachments {
		if value.VolumeId != volID {
			continue
		}

		//if the volume has been attached to this node, just return device attached before.
		if value.Host == hostname {
			//TODO: how to process status of volume to prevent volume concurrency.
			if value.Status != VOLUME_STATUS_ATTACHED {
				return Fail(errors.New("volume is processing."))
			}

			return Result{
				Status: "Success",
				Device: value.Mountpoint,
			}
		}
	}

	if len(attachments) > 0 {
		//if the volume has been attached to another node, and without MULTI capability, just return error.
		//TODO: should consider this case?
	}

	//create attachment to indicate the volume is been processed.
	iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}
	attachReq := &model.VolumeAttachmentSpec{
		VolumeId: volID,
		HostInfo: &model.HostInfo{
			Platform:  runtime.GOARCH,
			OsType:    runtime.GOOS,
			Ip:        iscsi.GetHostIp(),
			Host:      hostname,
			Initiator: localIqn,
		},
		Status: VOLUME_STATUS_ATTACHING,
	}
	attachSpec, errAttach := client.CreateVolumeAttachment(attachReq)
	if errAttach != nil {
		return Fail(errAttach.Error())
	}
	rollback := false
	defer func() {
		if rollback {
			client.DeleteVolumeAttachment(attachSpec.Id, attachSpec)
		}
	}()

	//as so far, only support iscsi protocol. if support multi protocols, the protocol type must be stored in volume,
	//otherwise can't get enough info from flexvolume framework.
	//iscsi implement as follow:
	iscsiCon := iscsi.ParseIscsiConnectInfo(attachSpec.ConnectionData)
	device, errConnect := iscsi.Connect(iscsiCon.TgtPortal, iscsiCon.TgtIQN, strconv.Itoa(iscsiCon.TgtLun))
	if errConnect != nil {
		rollback = true
		return Fail(errConnect.Error())
	}

	attachSpec.Status = VOLUME_STATUS_ATTACHED
	attachSpec.Mountpoint = device
	_, err = client.UpdateVolumeAttachment(attachSpec.Id, attachSpec)
	if err != nil {
		return Fail(err.Error())
	} else {
		return Result{
			Status: "Success",
			Device: device,
		}
	}
}

func (plugin *OpenSDSPlugin) Detach(volumeId string) Result {
	client := opensds.GetClient("")
	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return Fail(err.Error())
	}

	hostname, _ := os.Hostname()
	var act *model.VolumeAttachmentSpec = nil
	for _, value := range attachments {
		if hostname == value.Host && volumeId == value.VolumeId {
			//volume has attach to this node
			act = value
			break
		}
	}
	if act == nil {
		return Succeed()
	}

	if act.Mountpoint != "" {
		iscsiCon := iscsi.ParseIscsiConnectInfo(act.ConnectionData)
		err = iscsi.Disconnect(iscsiCon.TgtPortal, iscsiCon.TgtIQN)
		if err != nil {
			return Fail(err.Error())
		}
	}

	act.Status = VOLUME_STATUS_DETACHED
	err = client.DeleteVolumeAttachment(act.Id, act)
	if err != nil {
		return Fail(err.Error())
	}

	return Succeed()
}

func (plugin *OpenSDSPlugin) Mount(mountDir string, device string, opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)

	_, err := volume.MountVolume(mountDir, device, opt.FsType)
	if err != nil {
		return Fail(err.Error())
	}
	return Succeed()
}

func (plugin *OpenSDSPlugin) Unmount(mountDir string) Result {
	_, err := volume.UnmountVolume(mountDir)
	if err != nil {
		return Fail(err.Error())
	}
	return Succeed()
}

func main() {
	RunPlugin(&OpenSDSPlugin{})
}

