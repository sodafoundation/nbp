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

package org.opensds.vmware.ngc.adapters.oceanstor;

import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.List;

import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.entity.StringEntity;
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONArray;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.models.*;

class ProductModel {
    private static final HashMap<String, String> MODEL_MAP = new HashMap<>();

    static {
        MODEL_MAP.put("61", "6800 V3");
        MODEL_MAP.put("62", "6900 V3");
        MODEL_MAP.put("63", "5600 V3");
        MODEL_MAP.put("64", "5800 V3");
        MODEL_MAP.put("68", "5500 V3");
        MODEL_MAP.put("69", "2600 V3");
        MODEL_MAP.put("70", "5300 V3");
        MODEL_MAP.put("71", "2800 V3");
        MODEL_MAP.put("72", "18500 V3");
        MODEL_MAP.put("73", "18800 V3");
        MODEL_MAP.put("74", "2200 V3");
        MODEL_MAP.put("84", "2600F V3");
        MODEL_MAP.put("85", "5500F V3");
        MODEL_MAP.put("86", "5600F V3");
        MODEL_MAP.put("87", "5800F V3");
        MODEL_MAP.put("88", "6800F V3");
        MODEL_MAP.put("89", "18500F V3");
        MODEL_MAP.put("90", "18800F V3");
        MODEL_MAP.put("92", "2800 V5");
        MODEL_MAP.put("93", "5300 V5");
        MODEL_MAP.put("94", "5300F V5");
        MODEL_MAP.put("95", "5500 V5");
        MODEL_MAP.put("96", "5500F V5");
        MODEL_MAP.put("97", "5600 V5");
        MODEL_MAP.put("98", "5600F V5");
        MODEL_MAP.put("99", "5800 V5");
        MODEL_MAP.put("100", "5800F V5");
        MODEL_MAP.put("101", "6800 V5");
        MODEL_MAP.put("102", "6800F V5");
        MODEL_MAP.put("103", "18500 V5");
        MODEL_MAP.put("104", "18500F V5");
        MODEL_MAP.put("105", "18800 V5");
        MODEL_MAP.put("106", "18800F V5");
        MODEL_MAP.put("107", "5500 V5 Elite");
        MODEL_MAP.put("108", "2100 V3");
        MODEL_MAP.put("805", "Dorado5000 V3");
        MODEL_MAP.put("806", "Dorado6000 V3");
        MODEL_MAP.put("807", "Dorado18000 V3");
        MODEL_MAP.put("808", "Dorado NAS");
        MODEL_MAP.put("809", "Dorado NAS");
        MODEL_MAP.put("810", "Dorado3000 V3");
        MODEL_MAP.put("112", "2200 V3");
        MODEL_MAP.put("113", "2600 V3");
        MODEL_MAP.put("114", "2600F V3");
        MODEL_MAP.put("115", "5300 V5");
        MODEL_MAP.put("116", "5110 V5");
        MODEL_MAP.put("117", "5110F V5");
        MODEL_MAP.put("118", "5210 V5");
        MODEL_MAP.put("119", "5210F V5");
        MODEL_MAP.put("120", "5310 V5");
        MODEL_MAP.put("121", "5310F V5");
        MODEL_MAP.put("122", "5510 V5");
        MODEL_MAP.put("123", "5510F V5");
        MODEL_MAP.put("124", "5610 V5");
        MODEL_MAP.put("125", "5610F V5");
        MODEL_MAP.put("126", "5810 V5");
        MODEL_MAP.put("127", "5810F V5");
        MODEL_MAP.put("128", "6810 V5");
        MODEL_MAP.put("129", "6810F V5");
        MODEL_MAP.put("130", "18510 V5");
        MODEL_MAP.put("131", "18510F V5");
        MODEL_MAP.put("132", "18810 V5");
        MODEL_MAP.put("133", "18810F V5");
        MODEL_MAP.put("134", "5210 V5 Enhanced");
        MODEL_MAP.put("135", "5210F V5 Enhanced");
    }

    static String getModel(String model) {
        if (!MODEL_MAP.containsKey(model)) {
            return "unknown";
        }

        return (String) MODEL_MAP.get(model);
    }
}

class RunningStatus {
    private static final HashMap<String, String> RUNNINGSTATUS_MAP = new HashMap<>();

    static {
        RUNNINGSTATUS_MAP.put("1", "normal");
        RUNNINGSTATUS_MAP.put("3", "not running");
        RUNNINGSTATUS_MAP.put("12", "powering on");
        RUNNINGSTATUS_MAP.put("47", "powering off");
        RUNNINGSTATUS_MAP.put("51", "upgrading");
    }

    static String getStatus(String status) {
        if (!RUNNINGSTATUS_MAP.containsKey(status)) {
            return "unknown";
        }

        return RUNNINGSTATUS_MAP.get(status);
    }
}

class HealthStatus {
    private static final HashMap<String, String> HEALTHSTATUS_MAP = new HashMap<>();

    static {
        HEALTHSTATUS_MAP.put("1", "normal");
        HEALTHSTATUS_MAP.put("2", "fault");
    }

    static String getStatus(String status) {
        if (!HEALTHSTATUS_MAP.containsKey(status)) {
            return "unknown";
        }

        return HEALTHSTATUS_MAP.get(status);
    }
}

class VolumeMOBuilder {
    static public VolumeMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("NAME");
        String id = jsonObject.getString("ID");
        String wwn = jsonObject.getString("WWN");
        ALLOC_TYPE allocType = (jsonObject.getInt("ALLOCTYPE") == 1) ? ALLOC_TYPE.THIN : ALLOC_TYPE.THICK;
        long capacity = jsonObject.getLong("CAPACITY") * 512;
        long allocCapacity = jsonObject.getLong("ALLOCCAPACITY") * 512;
        VolumeMO.StatusE status = (jsonObject.getInt("HEALTHSTATUS") == 1) ? VolumeMO.StatusE.Normal : VolumeMO
                .StatusE.Faulty;
        String storagePoolId = jsonObject.getString("PARENTID");
        VolumeMO volumeMO = new VolumeMO(name, id, wwn, allocType, capacity);
        volumeMO.status = status;
        volumeMO.allocCapacity = allocCapacity;
        volumeMO.storagePoolId = storagePoolId;
        return volumeMO;
    }
}

class StoragePoolMOBuilder {
    static public StoragePoolMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("NAME");
        String id = jsonObject.getString("ID");
        POOL_TYPE type = (jsonObject.getInt("USAGETYPE") == 1) ? POOL_TYPE.BLOCK : POOL_TYPE.FILE;
        long totalCapacity = jsonObject.getLong("USERTOTALCAPACITY") * 512;
        long freeCapacity = jsonObject.getLong("USERFREECAPACITY") * 512;

        return new StoragePoolMO(name, id, type, totalCapacity, freeCapacity);
    }
}

class SnapshotMOBuilder {
    static public SnapshotMO build(JSONObject jsonObject) {
        String name = jsonObject.getString("NAME");
        String id = jsonObject.getString("ID");
        String healthStatus = HealthStatus.getStatus(jsonObject.getString("HEALTHSTATUS"));
        long capacity = jsonObject.getLong("USERCAPACITY") * 512;
        String parentId = jsonObject.getString("PARENTID");

        long timeStamp = jsonObject.getLong("TIMESTAMP");
        String activatedTime;
        if (timeStamp < 0) {
            activatedTime = "--";
        } else {
            Date date = new Date(timeStamp * 1000L);
            activatedTime = date.toString();
        }

        return new SnapshotMO(name, id, healthStatus, capacity, parentId, activatedTime);
    }
}

public class OceanStor extends Storage {
    class Handler implements Request.RequestHandler {
        @Override
        public void setRequestBody(HttpEntityEnclosingRequestBase req, Object body) {
            StringEntity entity = new StringEntity(body.toString(), "utf-8");
            req.setEntity(entity);
        }

        @Override
        public Object parseResponseBody(String body) throws JSONException {
            return new JSONObject(body);
        }
    }

    private RestClientWrapper client;

    public OceanStor(String name) {
        super(name);
        this.client = new RestClientWrapper();
    }

    @Override
    public void login(String ip, int port, String user, String password) throws Exception {
        client.login(ip, port, user, password);
    }

    @Override
    public void logout() {
        client.logout();
    }

    @Override
    public StorageMO getDeviceInfo() throws Exception {
        JSONObject system = client.getSystem();

        String name = system.getString("NAME");
        String model = "unknown";
        if (system.has("productModeString")) {
            model = system.getString("productModeString");
        } else if (system.has("PRODUCTMODE")) {
            model = ProductModel.getModel(system.getString("PRODUCTMODE"));
        }

        String sn = system.getString("wwn");
        String status = RunningStatus.getStatus(system.getString("RUNNINGSTATUS"));

        return new StorageMO(name, model, sn, status, "Huawei");
    }

    @Override
    public VolumeMO createVolume(String name, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception {
        JSONObject volume = client.createVolume(name, allocType, capacity, poolId);
        return VolumeMOBuilder.build(volume);
    }

    @Override
    public void deleteVolume(String volumeId) throws Exception {
        client.deleteVolume(volumeId);
    }

    @Override
    public List<VolumeMO> listVolumes() throws Exception {
        List<VolumeMO> volumes = new ArrayList<>();

        JSONArray jsonArray = client.listVolumes("");
        for (int i = 0; i < jsonArray.length(); i++) {
            JSONObject volume = jsonArray.getJSONObject(i);
            volumes.add(VolumeMOBuilder.build(volume));
        }

        return volumes;
    }

    @Override
    public List<VolumeMO> listVolumes(String poolId) throws Exception {
        List<VolumeMO> volumes = new ArrayList<>();

        JSONArray jsonArray = client.listVolumes(poolId);
        for (int i = 0; i < jsonArray.length(); i++) {
            JSONObject volume = jsonArray.getJSONObject(i);
            volumes.add(VolumeMOBuilder.build(volume));
        }

        return volumes;
    }

    @Override
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

    @Override
    public StoragePoolMO getStoragePool(String poolId) throws Exception {
        JSONObject pool = client.getStoragePool(poolId);
        return StoragePoolMOBuilder.build(pool);
    }

    @Override
    public VolumeMO queryVolumeByID(String volumeId) throws Exception {
        String subLunId = volumeId.substring(volumeId.length() - 8);
        Long numLunId = Long.parseLong(subLunId, 16);
        JSONObject volume = client.getVolumeById(String.valueOf(numLunId));
        return VolumeMOBuilder.build(volume);
    }

    @Override
    public List<SnapshotMO> listSnapshot(String volumeId) throws Exception {
        List<SnapshotMO> snapshots = new ArrayList<>();

        JSONArray jsonArray = client.listSnapshots(volumeId);
        for (int i = 0; i < jsonArray.length(); i++) {
            JSONObject snapshot = jsonArray.getJSONObject(i);
            snapshots.add(SnapshotMOBuilder.build(snapshot));
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
        client.rollbackVolumeSnapshot(snapshotId, rollbackSpeed);
    }
	
	@Override
    public void expandVolume(String volumeId, long capacity) throws Exception {
        client.expandVolume(volumeId, capacity);
    }


    private JSONObject getISCSIInitiator(String initiator) throws Exception {
        JSONObject jsonObject = client.getISCSIInitiator(initiator);
        if (jsonObject == null) {
            String msg = String.format("ISCSI initiator %s doesn't exist.", initiator);
            throw new Exception(msg);
        }

        if (jsonObject.getInt("RUNNINGSTATUS") != RUNNING_STATUS.ONLINE.getValue()) {
            String msg = String.format("ISCSI initiator %s isn't online.", initiator);
            throw new Exception(msg);
        }

        return jsonObject;
    }

    private List getFCInitiators(String[] initiators) throws Exception {
        List iniList = new ArrayList<JSONObject>();

        for (String i : initiators) {
            JSONObject jsonObject = client.getFCInitiator(i);
            if (jsonObject == null) {
                continue;
            }

            if (jsonObject.getInt("RUNNINGSTATUS") != RUNNING_STATUS.ONLINE.getValue()) {
                continue;
            }

            iniList.add(jsonObject);
        }

        if (iniList.isEmpty()) {
            String msg = String.format("No any of FC initiators %s exist and online.", initiators.toString());
            throw new Exception(msg);
        }

        return iniList;
    }

    private JSONObject getHost(ConnectMO connect, boolean onlyQuery) throws Exception {
        JSONObject iscsiInitiator = null;
        List fcInitiators = null;

        switch (connect.attachProtocol) {
            case ISCSI:
                iscsiInitiator = getISCSIInitiator(connect.iscsiInitiator);
                break;
            case FC:
                fcInitiators = getFCInitiators(connect.fcInitiators);
                break;
            default:
                try {
                    iscsiInitiator = getISCSIInitiator(connect.iscsiInitiator);
                } catch (Exception e) {
                    fcInitiators = getFCInitiators(connect.fcInitiators);
                }
                break;
        }

        String hostId = null;

        if (iscsiInitiator != null) {
            if (!iscsiInitiator.getBoolean("ISFREE")) {
                hostId = iscsiInitiator.getString("PARENTID");
            }
        } else if (fcInitiators != null) {
            for (Object i : fcInitiators) {
                JSONObject jsonObject = (JSONObject) i;
                if (!jsonObject.getBoolean("ISFREE")) {
                    hostId = jsonObject.getString("PARENTID");
                    break;
                }
            }
        }

        if (hostId != null) {
            return client.getHostById(hostId);
        }

        JSONObject host = client.getHostByName(connect.name);

        if (onlyQuery) {
            return host;
        }

        if (host == null) {
            host = client.createHost(connect.name, connect.osType);
        }

        if (iscsiInitiator != null) {
            client.addISCSIInitiatorToHost(iscsiInitiator.getString("ID"), host.getString("ID"));
        } else if (fcInitiators != null) {
            for (Object i : fcInitiators) {
                JSONObject jsonObject = (JSONObject) i;
                client.addFCInitiatorToHost(jsonObject.getString("ID"), host.getString("ID"));
            }
        }

        return host;
    }

    private JSONObject getHostGroup(JSONObject host, ConnectMO connect, boolean onlyQuery) throws Exception {
        JSONObject hostGroup = null;
        JSONArray hostGroups = client.getHostGroupsByHost(host.getString("ID"));

        if (hostGroups != null) {
            for (Object i : hostGroups) {
                JSONObject jsonObject = (JSONObject) i;
                JSONArray hosts = client.getHostsByHostGroup(jsonObject.getString("ID"));
                if (hosts != null && hosts.length() == 1) {
                    hostGroup = jsonObject;
                    break;
                }
            }
        }

        if (hostGroup != null) {
            return hostGroup;
        }

        hostGroup = client.getHostGroupByName(connect.name);
        if (onlyQuery) {
            return hostGroup;
        }

        if (hostGroup == null) {
            hostGroup = client.createHostGroup(connect.name);
        }

        client.addHostToHostGroup(host.getString("ID"), hostGroup.getString("ID"));

        return hostGroup;
    }

    private JSONObject getMappingView(JSONObject hostGroup, ConnectMO connect, boolean onlyQuery) throws Exception {
        JSONObject mappingView = null;

        if (hostGroup.getBoolean("ISADD2MAPPINGVIEW")) {
            JSONArray mappingViews = client.getMappingViewsByHostGroup(hostGroup.getString("ID"));
            if (mappingViews != null && !mappingViews.isEmpty()) {
                mappingView = mappingViews.getJSONObject(0);
            }
        }

        if (mappingView != null) {
            return mappingView;
        }

        mappingView = client.getMappingViewByName(connect.name);
        if (onlyQuery) {
            return mappingView;
        }

        if (mappingView == null) {
            mappingView = client.createMappingView(connect.name);
        }

        client.associateGroupToMappingView(
                hostGroup.getString("ID"), 14, mappingView.getString("ID"));

        return mappingView;
    }

    private JSONObject getLunGroup(JSONObject mappingView, ConnectMO connect, boolean onlyQuery) throws Exception {
        String mappingViewId = mappingView.getString("ID");
        JSONObject lunGroup = client.getLunGroupByMappingView(mappingViewId);
        if (lunGroup != null) {
            return lunGroup;
        }

        lunGroup = client.getLunGroupByName(connect.name);
        if (onlyQuery) {
            return lunGroup;
        }

        if (lunGroup == null) {
            lunGroup = client.createLunGroup(connect.name);
        }

        client.associateGroupToMappingView(
                lunGroup.getString("ID"), 256, mappingView.getString("ID"));

        return lunGroup;
    }

    private void addLunToLunGroup(String lunId, String lunGroupId) throws Exception {
        JSONArray lunGroups = client.getLunGroupsByLun(lunId);
        if (lunGroups != null) {
            for (Object i : lunGroups) {
                JSONObject jsonObject = (JSONObject) i;
                if (jsonObject.getString("ID").equals(lunGroupId)) {
                    return;
                }
            }
        }

        client.addLunToLunGroup(lunId, lunGroupId);
    }

    public void attachVolume(String volumeId, ConnectMO connect) throws Exception {
        JSONObject host = getHost(connect, false);

        JSONObject hostGroup = getHostGroup(host, connect, false);
        JSONObject mappingView = getMappingView(hostGroup, connect, false);
        JSONObject lunGroup = getLunGroup(mappingView, connect, false);

        addLunToLunGroup(volumeId, lunGroup.getString("ID"));
    }

    public void detachVolume(String volumeId, ConnectMO connect) throws Exception {
        JSONArray hosts = client.getHostsByLun(volumeId);
        if (hosts == null) {
            // Volume doesn't map to any host, directly return.
            return;
        }

        JSONObject host = getHost(connect, true);
        if (host == null) {
            return;
        }

        int i = 0;
        for (; i < hosts.length(); i++) {
            JSONObject jsonObject = hosts.getJSONObject(i);
            if (jsonObject.getString("NAME").equals(host.getString("NAME"))) {
                break;
            }
        }

        if (i >= hosts.length()) {
            return;
        }

        JSONObject hostGroup = getHostGroup(host, connect, true);
        if (hostGroup == null) {
            return;
        }

        JSONObject mappingView = getMappingView(hostGroup, connect, true);
        if (mappingView == null) {
            return;
        }

        JSONObject lunGroup = getLunGroup(mappingView, connect, true);
        if (lunGroup == null) {
            return;
        }

        client.removeLunFromLunGroup(volumeId, lunGroup.getString("ID"));
    }
}
