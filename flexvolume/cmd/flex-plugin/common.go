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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Result struct {
	// Status of the callout. One of "Success", "Failure" or "Not supported".
	Status string `json:"status"`
	// Reason for success/failure.
	Message string `json:"message,omitempty"`
	// Path to the device attached. This field is valid only for attach calls.
	// ie: /dev/sdx
	DevicePath string `json:"device,omitempty"`
	// Cluster wide unique name of the volume.
	VolumeName string `json:"volumeName,omitempty"`
	// Represents volume is attached on the node
	Attached bool `json:"attached,omitempty"`
	// Returns capabilities of the driver.
	// By default we assume all the capabilities are supported.
	// If the plugin does not support a capability, it can return false for that capability.
	Capabilities *DriverCapabilities `json:",omitempty"`
}

type DriverCapabilities struct {
	Attach         bool `json:"attach"`
	SELinuxRelabel bool `json:"selinuxRelabel"`
}

type FlexVolumePlugin interface {
	NewOptions() interface{}
	Init() Result
	GetVolumeName(opt interface{}) Result
	Attach(opt interface{}) Result
	Detach(volumeId string) Result
	WaitForDetach(device string) Result
	IsAttached(opt interface{}) Result
	WaitForAttach(device string, opt interface{}) Result
	MountDevice(mountDir string, device string, opt interface{}) Result
	UnmountDevice(mountDir string) Result
	Mount(mountDir string, opt interface{}) Result
	Unmount(mountDir string) Result
}

func Succeed(a ...interface{}) Result {
	return Result{
		Status:  "Success",
		Message: fmt.Sprint(a...),
	}
}

func Fail(a ...interface{}) Result {
	return Result{
		Status:  "Failure",
		Message: fmt.Sprint(a...),
	}
}

func finish(result Result) {
	code := 1
	if result.Status == "Success" {
		code = 0
	}

	var msg string
	res, err := json.Marshal(result)
	if err != nil {
		msg = "{\"status\":\"Failure\",\"message\":\"JSON error\"}"
		code = 1
	} else {
		msg = string(res)
	}

	fmt.Println(msg)
	log.Println(msg)
	os.Exit(code)
}

func usage() {
	cmd := os.Args[0]
	usage := `Invalid usage. Usage:
  %s init
  %s attach <json params> <nodename>
  %s detach <mount device> <nodename>
  %s waitforattach <mount device> <json params>
  %s mountdevice <mount dir> <mount device> <json params>
  %s unmountdevice <mount dir>
  %s isattached <json params> <nodename>
`
	fmt.Printf(usage, cmd, cmd, cmd, cmd, cmd, cmd, cmd)
	os.Exit(1)
}

func RunPlugin(plugin FlexVolumePlugin) {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "init":
		finish(plugin.Init())

	case "getvolumename":
		if len(os.Args) < 3 {
			finish(Fail("attach expected at least 3 arguments; got ", os.Args))
		}
		opt := plugin.NewOptions()
		if err := json.Unmarshal([]byte(os.Args[2]), opt); err != nil {
			finish(Fail("Could not parse options for getvolumename:", err))
		}
		finish(plugin.GetVolumeName(opt))

	case "attach":
		if len(os.Args) < 3 {
			finish(Fail("attach expected at least 3 arguments; got ", os.Args))
		}

		opt := plugin.NewOptions()
		if err := json.Unmarshal([]byte(os.Args[2]), opt); err != nil {
			finish(Fail("Could not parse options for attach:", err))
		}

		finish(plugin.Attach(opt))

	case "detach":
		if len(os.Args) < 3 {
			finish(Fail("detach expected at least 3 arguments; got ", os.Args))
		}

		volumeId := os.Args[2]
		finish(plugin.Detach(volumeId))

	case "waitForDetachCmd":
		if len(os.Args) != 3 {
			finish(Fail("waitForDetachCmd expected 3 arguments; got ", os.Args))
		}

		device := os.Args[2]
		finish(plugin.WaitForDetach(device))

	case "isattached":
		if len(os.Args) != 4 {
			finish(Fail("isattached expected 4 arguments; got ", os.Args))
		}

		opt := plugin.NewOptions()
		if err := json.Unmarshal([]byte(os.Args[2]), opt); err != nil {
			finish(Fail("Could not parse options for attach:", err))
		}

		finish(plugin.IsAttached(opt))

	case "waitforattach":
		if len(os.Args) != 4 {
			finish(Fail("waitforattach expected 4 arguments; got ", os.Args))
		}

		device := os.Args[2]

		opt := plugin.NewOptions()
		if err := json.Unmarshal([]byte(os.Args[3]), opt); err != nil {
			finish(Fail("Could not parse options for attach:", err))
		}

		finish(plugin.WaitForAttach(device, opt))

	case "mountdevice":
		if len(os.Args) != 5 {
			finish(Fail("mount device expected exactly 5 argument; got ", os.Args))
		}

		mountDir := os.Args[2]
		device := os.Args[3]

		opt := plugin.NewOptions()
		if err := json.Unmarshal([]byte(os.Args[4]), opt); err != nil {
			finish(Fail("Could not parse options for mount; got ", os.Args[3]))
		}

		finish(plugin.MountDevice(mountDir, device, opt))

	case "unmountdevice":
		if len(os.Args) != 3 {
			finish(Fail("unmount device expected exactly 3 argument; got ", os.Args))
		}

		mountDir := os.Args[2]

		finish(plugin.UnmountDevice(mountDir))

	case "mount":
		if len(os.Args) != 4 {
			finish(Fail("mount expected exactly 4 argument; got ", os.Args))
		}

		mountDir := os.Args[2]

		opt := plugin.NewOptions()
		if err := json.Unmarshal([]byte(os.Args[3]), opt); err != nil {
			finish(Fail("Could not parse options for mount; got ", os.Args[3]))
		}

		finish(plugin.Mount(mountDir, opt))

	case "unmount":
		if len(os.Args) != 3 {
			finish(Fail("mount expected exactly 3 argument; got ", os.Args))
		}

		mountDir := os.Args[2]

		finish(plugin.Unmount(mountDir))

	default:
		finish(Fail("Not sure what to do. Called with: ", os.Args))
	}
}

func FindLinkPath(device string) (string, error) {
	listCmd := exec.Command("ls", "/dev/disk/by-id", "-l")
	listCmdOut, err := listCmd.CombinedOutput()
	if err != nil {
		log.Printf("Could not find link path %s: %v\n", string(listCmdOut), err)
		return "", err
	}

	split := strings.Split(string(listCmdOut), "\n")

	var isFound, findStr = false, ""
	for i := 0; i < len(split); i++ {
		if strings.Contains(split[i], device) {
			isFound, findStr = true, split[i]
			break
		}
	}
	if !isFound {
		return "", errors.New("No link path matched!")
	}

	split1 := strings.Fields(findStr)
	volId := split1[8]
	linkPath := "/dev/disk/by-id/" + volId
	return linkPath, nil
}
