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
import java.util.List;

import org.json.JSONArray;
import org.json.JSONObject;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.models.*;

class VolumeMOBuilder {
    static public VolumeMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("name");
        String id = jsonObject.getString("id");
        JSONObject metadata = jsonObject.getJSONObject("metadata");
        String wwn = metadata.getString("wwn");
        ALLOC_TYPE allocType = ALLOC_TYPE.THIN;
        long capacity = jsonObject.getLong("size")*UNIT_TYPE.GB.getUnit();
        long allocCapacity = jsonObject.getLong("size") * UNIT_TYPE.GB.getUnit();
        String volStatus = jsonObject.getString("status");
        VolumeMO.StatusE status = (volStatus.equals("available") ||
                volStatus.equals("inUse")) ? VolumeMO.StatusE.Normal : VolumeMO
                .StatusE.Faulty;
        String storagePoolId = jsonObject.getString("poolId");
        VolumeMO volumeMO = new VolumeMO(name, id, wwn, allocType, capacity);
        volumeMO.status = status;
        volumeMO.allocCapacity = allocCapacity;
        volumeMO.storagePoolId = storagePoolId;
        return volumeMO;
    }
}

class StoragePoolMOBuilder {
    static public StoragePoolMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("name");
        String id = jsonObject.getString("id");
        POOL_TYPE type = (jsonObject.getString("storageType").equals("block")) ? POOL_TYPE.BLOCK : POOL_TYPE.FILE;
        long totalCapacity = jsonObject.getLong("totalCapacity")*UNIT_TYPE.GB.getUnit();
        long freeCapacity = jsonObject.getLong("freeCapacity")*UNIT_TYPE.GB.getUnit();

        return new StoragePoolMO(name, id, type, totalCapacity, freeCapacity);
    }
}

class SnapshotMOBuilder {
    static public SnapshotMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("name");
        String id = jsonObject.getString("id");
        String healthStatus = jsonObject.getString("status");
        long capacity = jsonObject.getLong("size")*UNIT_TYPE.GB.getUnit();
        String parentId = jsonObject.getString("volumeId");
        String timestamp = jsonObject.getString("createdAt");

        return new SnapshotMO(name, id, healthStatus, capacity, parentId, timestamp);
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
        return client.getDeviceInfo();
    }

    public VolumeMO createVolume(String name, String description, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception {
        // convert capacity from Bytes to GB
    	capacity = capacity/(UNIT_TYPE.GB.getUnit());
        JSONObject volume = client.createVolume(name, description, allocType, capacity, poolId);
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

    @Override
	public List<VolumeMO> listVolumes(String filterKey, String filtervalue) throws Exception {
		 List<VolumeMO> volumes = new ArrayList<>();

	        JSONArray jsonArray = client.listVolumes(filterKey, filtervalue);
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
        client.attachVolume(volumeId, connect.iscsiInitiator, connect.initiatorIp);
    }

    public void detachVolume(String volumeId, ConnectMO connect) throws Exception {
        client.detachVolume(volumeId);
    }
	
	
    @Override
    public VolumeMO queryVolumeByID(String volumeId) throws Exception {
        JSONObject volume = client.getVolume(volumeId);
        return VolumeMOBuilder.build(volume);
    }

    @Override
    public List<SnapshotMO> listSnapshot(String volumeId) throws Exception {
         List<SnapshotMO> snapshots = new ArrayList<>();

	     JSONArray jsonArray = client.listVolumeSnapshot(volumeId);
	     if (jsonArray != null) {
	          for (int i = 0; i < jsonArray.length(); i++) {
	              JSONObject snapshot = jsonArray.getJSONObject(i);
	              snapshots.add(SnapshotMOBuilder.build(snapshot));
	          }
	      }

	    return snapshots;
    }

    @Override
    public void createVolumeSnapshot(String volumeId, String name) throws Exception {
        client.createVolumeSnapshot(volumeId, name);
    }

    @Override
    public void deleteVolumeSnapshot(String snapshotId) throws Exception {
        client.deleteVolumeSnapshot(snapshotId);
    }

    @Override
    public void rollbackVolumeSnapshot(String snapshotId, String rollbackSpeed) throws Exception {
         //TODO
    }
	
	@Override
    public void expandVolume(String volumeId, long capacity) throws Exception {
		capacity = capacity/(UNIT_TYPE.GB.getUnit());
		client.expandVolume(volumeId, capacity);
    }
}
