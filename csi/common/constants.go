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

package common

// K8s storage class parameter keywords
const (
	ParamProfile           = "profile"
	ParamEnableReplication = "enableReplication"
	ParamSecondaryAZ       = "secondaryAvailabilityZone"
)

// CSI publish attribute keywords
const (
	PublishHostIp            = "hostIp"
	PublishHostName          = "hostName"
	PublishAttachId          = "attachmentId"
	PublishSecondaryAttachId = "secondaryAttachmentId"
	PublishAttachStatus      = "attachmentStatus"
	PublishAttachMode        = "attachMode"
)

const (
	DefaultAvailabilityZone = "default"
)

// PluginName setting
const (
	FakeIQN = "fakeIqn"
)
