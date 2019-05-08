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

package main

import (
	"io/ioutil"
	"os"

	"encoding/json"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/opensds/nbp/csi/util"
)

const (
	dbPath                = "csi_client_db"
	volumeFileName        = "csi_client_db/volume"
	publishVolumeFileName = "csi_client_db/publish_volume"
	testStageFileName     = "csi_client_db/stage"
)

const (
	stageStart             = "start"
	stageVolume            = "volume"
	stageControllerPublish = "controllerPublish"
	stageNodePublish       = "nodePublish"
)

func init() {
	if exist, _ := util.PathExists(dbPath); !exist {
		os.MkdirAll(dbPath, 0755)
	}
}

func readFile(fname string) ([]byte, error) {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		glog.Error(err)
	}
	return b, err
}

func writeFile(fname string, data []byte) error {
	return ioutil.WriteFile(fname, data, 0644)
}

func StoreVolume(vol *csi.Volume) error {
	data, err := json.Marshal(vol)
	if err != nil {
		return err
	}
	return writeFile(volumeFileName, data)
}

func GetVolume() (*csi.Volume, error) {
	vol := &csi.Volume{}
	data, err := readFile(volumeFileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, vol)
	return vol, err
}

func StorePublishVolume(publishVol map[string]string) error {
	data, err := json.Marshal(publishVol)
	if err != nil {
		return err
	}
	return writeFile(publishVolumeFileName, data)
}

func GetPublishVolume() (map[string]string, error) {
	publishVol := map[string]string{}
	data, err := readFile(publishVolumeFileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &publishVol)
	return publishVol, err
}

func GetTestStage() (string, error) {
	data, err := readFile(testStageFileName)
	if err != nil {
		if !os.IsExist(err) {
			return stageStart, nil
		} else {
			return "", err
		}
	}
	return string(data), nil
}

func SetTestStage(stage string) error {
	err := writeFile(testStageFileName, []byte(stage))
	if err != nil {
		glog.Errorf("Set stage fail %v", err)
	}
	return err
}

func RemoveDbFile() {
	os.RemoveAll(dbPath)
}
