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

package org.opensds.vmware.ngc.adapters.opensds;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.entity.StringEntity;
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONTokener;
import org.json.JSONArray;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.models.*;


class VolumeMOBuilder {
    static public VolumeMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("name");
        String id = jsonObject.getString("id");
        String wwn = jsonObject.getString("id");
        ALLOC_TYPE allocType = ALLOC_TYPE.THIN;
        long capacity = jsonObject.getLong("size");

        return new VolumeMO(name, id, wwn, allocType, capacity);
    }
}

class StoragePoolMOBuilder {
    static public StoragePoolMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("name");
        String id = jsonObject.getString("id");
        POOL_TYPE type = (jsonObject.getString("storageType").equals("block")) ? POOL_TYPE.BLOCK : POOL_TYPE.FILE;
        long totalCapacity = jsonObject.getLong("totalCapacity");
        long freeCapacity = jsonObject.getLong("freeCapacity");

        return new StoragePoolMO(name, id, type, totalCapacity, freeCapacity);
    }
}

public class OpenSDS extends Storage {

    RestClient client;

    public OpenSDS(String name) {
        super(name);
        this.client = new RestClient();
    }

    public void login(String ip, int port, String user, String password) throws Exception {
        client.login(ip, port, user, password);
    }

    public void logout() {
        client.logout();
    }

    public StorageMO getDeviceInfo() throws Exception {
        return new StorageMO(name, "v1", "", "Available", "OpenSDS");
    }

    public VolumeMO createVolume(String name, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception {
        JSONObject volume = client.createVolume(name, allocType, capacity, poolId);
        return VolumeMOBuilder.build(volume);
    }

    public void deleteVolume(String volumeId) throws Exception {
        client.deleteVolume(volumeId);
    }

    public List<VolumeMO> listVolumes() throws Exception {
        List<VolumeMO> volumes = new ArrayList<>();

        JSONArray jsonArray = client.listVolumes("");
        if (jsonArray != null) {
            for (int i = 0; i < jsonArray.length(); i++) {
                JSONObject volume = jsonArray.getJSONObject(i);
                volumes.add(VolumeMOBuilder.build(volume));
            }
        }

        return volumes;
    }

    public List<VolumeMO> listVolumes(String poolId) throws Exception {
        List<VolumeMO> volumes = new ArrayList<>();

        JSONArray jsonArray = client.listVolumes(poolId);
        if (jsonArray != null) {
            for (int i = 0; i < jsonArray.length(); i++) {
                JSONObject volume = jsonArray.getJSONObject(i);
                volumes.add(VolumeMOBuilder.build(volume));
            }
        }

        return volumes;
    }

    public List<StoragePoolMO> listStoragePools() throws Exception {
        List<StoragePoolMO> pools = new ArrayList<>();

        JSONArray jsonArray = client.listStoragePools();
        if (jsonArray != null) {
            for (int i = 0; i < jsonArray.length(); i++) {
                JSONObject pool = jsonArray.getJSONObject(i);
                pools.add(StoragePoolMOBuilder.build(pool));
            }
        }

        return pools;
    }

    public StoragePoolMO getStoragePool(String poolId) throws Exception {
        JSONObject pool = client.getStoragePool(poolId);
        return StoragePoolMOBuilder.build(pool);
    }

    public void attachVolume(String volumeId, ConnectMO connect) throws Exception {
    }

    public void detachVolume(String volumeId, ConnectMO connect) throws Exception {
    }
}
