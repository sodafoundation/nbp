// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.
package org.opensds.vmware.ngc.model.datastore;

import com.vmware.vim25.HostScsiDiskPartition;
import org.opensds.vmware.ngc.model.VolumeInfo;

import java.util.ArrayList;
import java.util.List;

public class VMFSDatastore extends Datastore{

    private String uuid;

    private VolumeInfo[] volumeInfos;

    private String vmfsVersion;

    private boolean isLocal;

    private final List<HostScsiDiskPartition> hostScsiDiskPartitionList = new ArrayList<>();

    public List<HostScsiDiskPartition> getHostScsiDiskPartitionList() {
        return hostScsiDiskPartitionList;
    }

    public boolean isLocal() {
        return isLocal;
    }

    public void setLocal(boolean local) {
        isLocal = local;
    }

    public VolumeInfo[] getVolumeInfos() {
        return volumeInfos;
    }

    public String getVmfsVersion() {
        return vmfsVersion;
    }

    public void setVolumeInfos(VolumeInfo[] volumeInfos) {
        this.volumeInfos = volumeInfos;
    }

    public void setVmfsVersion(String vmfsVersion) {
        this.vmfsVersion = vmfsVersion;
    }

    public void setUuid(String uuid) {
        this.uuid = uuid;
    }

    public String getUuid() {
        return uuid;
    }
}
