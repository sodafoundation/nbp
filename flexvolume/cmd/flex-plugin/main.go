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

func (plugin *OpenSDSPlugin) MountDevice(mountDir string, device string, opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)

	act := getAttachmentByVolumeId(opt.VolumeId)
	if act == nil || len(act.Mountpoint) == 0 || act.Mountpoint != device {
		return Fail(errors.New("mount device is not exist"))
	}

	_, err := volume.MountVolume("", mountDir, device, opt.FsType, opt.AccessMode)
	if err != nil {
		return Fail(err.Error())
	}

	//save baseMountPath for Follow-up call
	//TODO: when unmount device, how to clean info? or select a mount path from /proc/mounts as baseMountPath?
	act.Metadata["baseMountPath"] = mountDir

	client := opensds.GetClient("")
	_, err = client.UpdateVolumeAttachment(act.Id, act)
	if err != nil {
		return Fail(err.Error())
	}
	return Succeed()
}

func (plugin *OpenSDSPlugin) UnmountDevice(mountDir string) Result {
	_, err := volume.UnmountVolume(mountDir)
	if err != nil {
		return Fail(err.Error())
	}
	return Succeed()
}

func (plugin *OpenSDSPlugin) Mount(mountDir string, opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)

	//find mount device
	act := getAttachmentByVolumeId(opt.VolumeId)
	if act == nil || len(act.Mountpoint) == 0 {
		return Fail(errors.New("mount device is not exist"))
	}

	//accord to flexvolume design, mount should exec "mount --bind" command which depends on baseMountPath,
	//so check whether baseMountPath is valid.
	if _, exist := act.Metadata["baseMountPath"]; !exist || len(act.Metadata["baseMountPath"]) == 0 {
		return Fail(errors.New("mount device failed"))
	}

	_, err := volume.MountVolume(act.Metadata["baseMountPath"], mountDir, act.Mountpoint, opt.FsType, opt.AccessMode)
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

func (plugin *OpenSDSPlugin) IsAttached(opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)
	act := getAttachmentByVolumeId(opt.VolumeId)
	if act == nil || len(act.Mountpoint) == 0 {
		return Result{
			Status:   "Success",
			Attached: false,
		}
	}

	return Result{
		Status:   "Success",
		Attached: true,
	}
}

func (plugin *OpenSDSPlugin) WaitForAttach(device string, opts interface{}) Result {
	result := plugin.Attach(opts)
	if result.Status != "Success" {
		return result
	}

	if len(device) != 0 && result.Device != device {
		return Fail(errors.New("the volume has attached another device."))
	}

	return result
}

func (plugin *OpenSDSPlugin) WaitForDetach(device string) Result {
	act := getAttachmentByDevice(device)
	if act == nil {
		return Succeed()
	}

	return plugin.Detach(act.VolumeId)
}

func getAttachmentByVolumeId(volumeId string) *model.VolumeAttachmentSpec {
	client := opensds.GetClient("")
	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return nil
	}

	hostname, _ := os.Hostname()
	var act *model.VolumeAttachmentSpec = nil
	for _, actValue := range attachments {
		if actValue.VolumeId == volumeId && actValue.Host == hostname {
			act = actValue
			break
		}
	}

	return act
}

func getAttachmentByDevice(device string) *model.VolumeAttachmentSpec {
	client := opensds.GetClient("")
	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return nil
	}

	hostname, _ := os.Hostname()
	var act *model.VolumeAttachmentSpec = nil
	for _, actValue := range attachments {
		//must ensure the device used by only one volume.
		if actValue.Mountpoint == device && actValue.Host == hostname {
			act = actValue
			break
		}
	}

	return act
}

func main() {
	RunPlugin(&OpenSDSPlugin{})
}
