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

package block

// CSI volume attribute keywords
const (
	VolumeName          = "name"
	VolumeStatus        = "status"
	VolumeAZ            = "availabilityZone"
	VolumePoolId        = "poolId"
	VolumeProfileId     = "profileId"
	VolumeLvPath        = "lvPath"
	VolumeReplicationId = "replicationId"
)

// Opensds replication metadata keywords
const (
	AttachedVolumeId = "attachedVolumeId"
	AttachedId       = "attachedId"
)

// Opensds Attachment metadata keywords
const (
	TargetPath        = "targetPath"
	StagingTargetPath = "stagingTargetPath"
)

// volume prefix
const SecondaryPrefix = "secondary-"

// Csi configuration parameters and values
const (
	// parameters
	CSIVolumeMode = "CSIVolumeMode"

	// CSIVolumeMode = Filesystem
	CSIFilesystem = "Filesystem"

	// CSIVolumeMode = Block
	CSIBlock = "Block"
)

// PluginName setting
const (
	PluginName = "csi-opensdsplugin-block"
)

var TopologyZoneKey = "topology." + PluginName + "/zone"
