// Copyright 2019 The OpenSDS Authors.
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

package block

import (
	"errors"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"strings"
)

// CreateSymlink Symlink implementation
func CreateSymlink(device, mountpoint string) error {
	_, err := os.Lstat(mountpoint)
	if err != nil && os.IsNotExist(err) {
		glog.V(5).Infof("mountpoint=%v does not exist", mountpoint)
	} else {
		glog.Errorf("mountpoint=%v already exists", mountpoint)
		// The mountpoint deleted here is a folder or a soft connection.
		// From the test results, this is fine.
		_, err := exec.Command("rm", "-rf", mountpoint).CombinedOutput()

		if nil != err {
			glog.Errorf("faild to delete %v", mountpoint)
			return err
		}
	}

	err = os.Symlink(device, mountpoint)
	if err != nil {
		glog.Errorf("failed to create a link: oldname=%v, newname=%v\n", device, mountpoint)
		return err
	}

	return nil
}

// ExtractISCSIInitiatorFromNodeInfo implementation
func ExtractISCSIInitiatorFromNodeInfo(nodeInfo string) (string, error) {
	for _, v := range strings.Split(nodeInfo, ",") {
		if strings.Contains(v, "iqn") {
			glog.V(5).Infof("ISCSI initiator is %s", v)
			return v, nil
		}
	}

	return "", errors.New("no ISCSI initiators found")
}

// ExtractNvmeofInitiatorFromNodeInfo implementation
func ExtractNvmeofInitiatorFromNodeInfo(nodeInfo string) (string, error) {
	for _, v := range strings.Split(nodeInfo, ",") {
		if strings.Contains(v, "nqn") {
			glog.V(5).Info("Nvmeof initiator is ", v)
			return v, nil
		}
	}

	return "", errors.New("no Nvmeof initiators found")
}

// ExtractFCInitiatorFromNodeInfo implementation
func ExtractFCInitiatorFromNodeInfo(nodeInfo string) ([]string, error) {
	var wwpns []string
	for _, v := range strings.Split(nodeInfo, ",") {
		if strings.Contains(v, "node_name") {
			wwpns = append(wwpns, strings.Split(v, ":")[1])
		}
	}

	if len(wwpns) == 0 {
		return nil, errors.New("no FC initiators found")
	}

	glog.V(5).Infof("FC initiators are %s", wwpns)

	return wwpns, nil
}
