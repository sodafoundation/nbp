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

package opensds

// K8s storage class parameter keywords
const (
	ParamProfile           = "profile"
	ParamEnableReplication = "enableReplication"
	ParamSecondaryAZ       = "secondaryAvailabilityZone"
)

// CSI volume attribute keywords
const (
	VolumeName          = "name"
	VolumeStatus        = "status"
	VolumeAZ            = "availabilityZone"
	VolumePoolId        = "poolId"
	VolumeProfileId     = "profileId"
	VolumeLvPath        = "lvPath"
	VolumeReplicationId = "replicationId"
	StorageType         = "storageType"
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

// Opensds Attachment metadata keywords
const (
	TargetPath        = "targetPath"
	StagingTargetPath = "stagingTargetPath"
)

// Opensds replication metadata keywords
const (
	AttachedVolumeId = "attachedVolumeId"
	AttachedId       = "attachedId"
)

// volume prefix
const SecondaryPrefix = "secondary-"

const (
	// default filesystem type
	DefFSType               = "ext4"
	DefaultAvailabilityZone = "default"
)

// Csi configuration parameters and values
const (
	// parameters
	CSIVolumeMode = "CSIVolumeMode"

	// CSIVolumeMode = Filesystem
	CSIFilesystem = "Filesystem"

	// CSIVolumeMode = Block
	CSIBlock = "Block"
)

// fileshare constant parameters
const (
	ShareName       = "shareName"
	ShareAZ         = "shareAZ"
	ShareStatus     = "shareStatus"
	SharePoolId     = "sharePoolId"
	ShareProfileId  = "shareProfileId"
	ShareProtocol   = "shareProtocol"
	NFS             = "nfs"
	IpIdx           = 2
	ExportLocations = "exportLocations"
	FileShareName   = "fileShareName"
)

// PluginName setting
const (
	FakeIQN    = "fakeIqn"
	PluginName = "csi-opensdsplugin"
)

var TopologyZoneKey = "topology." + PluginName + "/zone"
