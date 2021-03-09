// Copyright 2018 The OpenSDS Authors.
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

package file

import (
	"github.com/sodafoundation/api/client"
	"github.com/sodafoundation/nbp/csi/common"
)

// Plugin define
type Plugin struct {
	FileShareClient *FileShare
}

// NewServer initializes and return plugin server
func NewServer(client *client.Client) (common.Service, error) {
	p := &Plugin{
		FileShareClient: NewFileshare(client),
	}

	// When there are multiple volumes unmount at the same time,
	// it will cause conflicts related to the state machine,
	// so start a watch list to let the volumes unmount one by one.
	go common.UnpublishRoutine(client)

	return p, nil
}
