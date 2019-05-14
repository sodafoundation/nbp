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

package util

const (
	// NameSpace for CSI
	NameSpace = "csi"

	// CSI endpoint environment variable name
	CSIEndpoint = "CSI_ENDPOINT"
	// CSI default endpoint
	CSIDefaultEndpoint = "unix://path/to/unix/domain/socket.sock"

	// Opensds endpoint environment variable name
	OpensdsEndpoint = "OPENSDS_ENDPOINT"
	// Opensds default endpoint
	OpensdsDefaultEndpoint = "http://localhost:50040"

	// Opensds auth strategy
	OpensdsAuthStrategy        = "OPENSDS_AUTH_STRATEGY"
	OpensdsDefaultAuthStrategy = "noauth"

	//  Opensds Secondary AZ
	OpensdsSecondaryAZ        = "OPENSDS_SECONDARY_AZ"
	OpensdsDefaultSecondaryAZ = "secondary"

	// CSI  environment variable whether enable the replication function, value can be true or false
	CSIEnableReplication = "CSI_ENABLE_REPLICATION"

	// 1024 * 1024 * 1024
	GiB int64 = 1073741824
)
