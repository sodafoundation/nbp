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

package sanity

import (
	"github.com/opensds/opensds/contrib/connector"
)

type FakeMounter struct{}

func getFakeMounter() connector.Mounter {
	return &FakeMounter{}
}

func (*FakeMounter) GetFSType(device string) (string, error) {
	return "ext4", nil
}

func (*FakeMounter) Format(device string, fsType string) error {
	return nil
}

func (*FakeMounter) Mount(device, mountpoint, fsType string, mountFlags []string) error {
	return nil
}

func (*FakeMounter) Umount(mountpoint string) error {
	return nil
}

func (*FakeMounter) GetHostIP() string {
	return "127.0.0.1"
}

func (*FakeMounter) GetHostName() (string, error) {
	return "fake-host", nil
}

func (*FakeMounter) IsMounted(target string) (bool, error) {
	return false, nil
}
