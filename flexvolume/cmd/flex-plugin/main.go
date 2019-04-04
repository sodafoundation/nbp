// The MIT License (MIT)
// Copyright (c) 2016 Tony Zou
//
//    Permission is hereby granted, free of charge, to any person obtaining a
//    copy of this software and associated documentation files (the "Software"),
//    to deal in the Software without restriction, including without limitation
//    the rights to use, copy, modify, merge, publish, distribute, sub license,
//    and/or sell copies of the Software, and to permit persons to whom the
//    Software is furnished to do so, subject to the following conditions:
//
//    The above copyright notice and this permission notice shall be included
//    in all copies or substantial portions of the Software.
//
//    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
//    OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//    FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL
//    THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
//    FROM,OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
//    IN THE SOFTWARE.

//----------------------------------------------------------------------------
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
	"log"
	"os"
	"runtime"

	"github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/flexvolume/pkg/volume"
	"github.com/opensds/opensds/contrib/connector"
	_ "github.com/opensds/opensds/contrib/connector/rbd"
	"github.com/opensds/opensds/pkg/model"
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
	AccessMode string `json:"kubernetes.io/readwrite"`
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

func (plugin *OpenSDSPlugin) GetVolumeName(opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)
	volId := opt.VolumeId

	client, err := opensds.GetClient("", "")
	if err != nil {
		return Fail(errors.New(fmt.Sprintf("get client failed, %s", err.Error())))
	}
	vol, err := client.GetVolume(volId)

	if err != nil {
		return Fail(errors.New(fmt.Sprintf("volume not exist, %s", err.Error())))
	} else {
		return Result{
			Status:     "Success",
			VolumeName: vol.Name,
		}
	}
}

func (plugin *OpenSDSPlugin) Attach(opts interface{}) Result {
	opt := opts.(*OpenSDSOptions)
	volID := opt.VolumeId

	client, err := opensds.GetClient("", "")
	if err != nil {
		return Fail(errors.New(fmt.Sprintf("get client failed, %s", err.Error())))
	}

	vol, errVol := client.GetVolume(volID)
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
				Status:     "Success",
				DevicePath: value.Mountpoint,
			}
		}
	}

	if len(attachments) > 0 {
		//if the volume has been attached to another node, and without MULTI capability, just return error.
		//TODO: should consider this case?
	}

	//create attachment to indicate the volume is been processed.
	localIqn, _ := connector.NewConnector("iscsi").GetInitiatorInfo()

	attachReq := &model.VolumeAttachmentSpec{
		VolumeId: volID,
		HostInfo: model.HostInfo{
			Platform:  runtime.GOARCH,
			OsType:    runtime.GOOS,
			Ip:        connector.GetHostIP(),
			Host:      hostname,
			Initiator: localIqn,
		},
		Status:   VOLUME_STATUS_ATTACHING,
		Metadata: vol.Metadata,
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

	volConnector := connector.NewConnector(attachSpec.DriverVolumeType)
	if volConnector == nil {
		rollback = true
		return Fail(errors.New(fmt.Sprintf("Unsupport driverVolumeType: %s", attachSpec.DriverVolumeType)))
	}

	device, errAttach := volConnector.Attach(attachSpec.ConnectionData)
	if errAttach != nil {
		rollback = true
		return Fail(errAttach.Error())
	}

	attachSpec.Status = VOLUME_STATUS_ATTACHED
	attachSpec.Mountpoint = device
	_, err = client.UpdateVolumeAttachment(attachSpec.Id, attachSpec)
	if err != nil {
		return Fail(err.Error())
	} else {
		return Result{
			Status:     "Success",
			DevicePath: device,
		}
	}
}

func (plugin *OpenSDSPlugin) Detach(volumeId string) Result {
	client, err := opensds.GetClient("", "")
	if err != nil {
		return Fail(errors.New(fmt.Sprintf("get client failed, %s", err.Error())))
	}

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
		volConnector := connector.NewConnector(act.DriverVolumeType)
		if volConnector == nil {
			return Fail(errors.New(fmt.Sprintf("Unsupport driverVolumeType: %s", act.DriverVolumeType)))
		}

		err = volConnector.Detach(act.ConnectionData)
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
	if act.Metadata == nil {
		act.Metadata = map[string]string{}
	}
	act.Metadata["baseMountPath"] = mountDir

	client, err := opensds.GetClient("", "")
	if err != nil {
		return Fail(errors.New(fmt.Sprintf("get client failed, %s", err.Error())))
	}

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

	if len(device) != 0 && result.DevicePath != device {
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
	client, err := opensds.GetClient("", "")
	if err != nil {
		return nil
	}

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
	client, err := opensds.GetClient("", "")
	if err != nil {
		return nil
	}

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

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func main() {
	if exist, _ := pathExists("/var/log/opensds"); !exist {
		os.MkdirAll("/var/log/opensds", 0755)
	}
	// Open OpenSDS flexvolume service log file
	f, err := os.OpenFile("/var/log/opensds/flexvolume.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Errorf("Error opening file:%v", err)
		os.Exit(1)
	}
	defer f.Close()

	// assign it to the standard logger
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Cmd:", os.Args)
	RunPlugin(&OpenSDSPlugin{})
}
