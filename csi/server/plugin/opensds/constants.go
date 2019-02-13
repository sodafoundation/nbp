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

package opensds

// K8s storage class parameter keywords

const (
	KParamProfile           = "profile"
	KParamAZ                = "availabilityzone"
	KParamEnableReplication = "enablereplication"
	KParamSecondaryAZ       = "secondaryavailabilityzone"
)

// CSI volume attribute keywords
const (
	KVolumeName          = "name"
	KVolumeStatus        = "status"
	KVolumeAZ            = "availabilityZone"
	KVolumePoolId        = "poolId"
	KVolumeProfileId     = "profileId"
	KVolumeLvPath        = "lvPath"
	KVolumeReplicationId = "replicationId"
)

// CSI publish attribute keywords
const (
	KPublishHostIp            = "HostIp"
	KPublishHostName          = "HostName"
	KPublishAttachId          = "AttachmentId"
	KPublishSecondaryAttachId = "SecondaryAttachmentId"
	KPublishAttachStatus      = "AttachmentStatus"
)

// Opensds Attachment metadata keywords
const (
	KTargetPath        = "targetPath"
	KStagingTargetPath = "stagingTargetPath"
)

// Opensds replication metadata keywords
const (
	KAttachedVolumeId = "attachedVolumeId"
	KAttachedId       = "attachedId"
)

// volume prefix
const SecondaryPrefix = "secondary-"

const (
	// DefFSType default filesystem type
	DefFSType = "ext4"
)

// Csi configuration parameters and values
const (
	// parameters
	KCSIVolumeMode = "CSIVolumeMode"

	// CSIVolumeMode = Filesystem
	KCSIFilesystem = "Filesystem"

	// CSIVolumeMode = Block
	KCSIBlock = "Block"
)
