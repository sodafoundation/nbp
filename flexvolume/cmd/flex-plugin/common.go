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
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Device   string `json:"device,omitempty"`
	Attached bool   `json:"attached,omitempty"`
}

type FlexVolumePlugin interface {
	NewOptions() interface{}
	Init() Result
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
	res, err := json.Marshal(result)
	if err != nil {
		fmt.Println("{\"status\":\"Failure\",\"message\":\"JSON error\"}")
	} else {
		fmt.Println(string(res))
	}
	os.Exit(code)
}

func RunPlugin(plugin FlexVolumePlugin) {
	if len(os.Args) < 2 {
		finish(Fail("Expected at least one argument"))
	}

	switch os.Args[1] {
	case "init":
		finish(plugin.Init())

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
		log.Printf("Could not find link path %f: %v\n", listCmdOut, err)
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
