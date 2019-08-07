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

package org.opensds.vmware.ngc.common;

import org.opensds.vmware.ngc.models.*;

import java.util.List;

public abstract class Storage {
    protected String name;

    public Storage(String name) {
        this.name = name;
    }

    public abstract void login(String ip, int port, String user, String password) throws Exception;
    public abstract void logout();
    public abstract StorageMO getDeviceInfo() throws Exception;
    public abstract VolumeMO createVolume(String name, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception;
    public abstract void deleteVolume(String volumeId) throws Exception;
    public abstract List<VolumeMO> listVolumes() throws Exception;
    public abstract List<VolumeMO> listVolumes(String poolId) throws Exception;
    public abstract List<StoragePoolMO> listStoragePools() throws Exception;
    public abstract StoragePoolMO getStoragePool(String poolId) throws Exception;
    public abstract void attachVolume(String volumeId, ConnectMO connect) throws Exception;
    public abstract void detachVolume(String volumeId, ConnectMO connect) throws Exception;
}
