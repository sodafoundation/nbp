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

import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.entity.StringEntity;
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONArray;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.exceptions.NotAuthorizedException;
import org.opensds.vmware.ngc.models.ALLOC_TYPE;
import org.opensds.vmware.ngc.models.HOST_OS_TYPE;

import java.net.URLEncoder;

class RestClient {
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

    private Request request;

    private long getErrorCode(JSONObject response) {
        JSONObject error = response.getJSONObject("error");
        return error.getLong("code");
    }

    private String getErrorDescription(JSONObject response) {
        JSONObject error = response.getJSONObject("error");
        return error.getString("description");
    }

    private boolean fail(JSONObject response) throws Exception {
        long errorCode = getErrorCode(response);
        if (errorCode == -401) {
            throw new NotAuthorizedException(getErrorDescription(response));
        }

        return errorCode != 0;
    }

    void login(String ip, int port, String user, String password) throws Exception {
        if (this.request != null) {
            this.request.close();
        }

        Request request = new Request(ip, port, new Handler());
        request.setHeaders("Content-Type", "application/json");

        JSONObject requestData = new JSONObject();
        requestData.put("username", user);
        requestData.put("password", password);
        requestData.put("scope", "0");

        JSONObject response = (JSONObject)request.post(
                "/deviceManager/rest/xxxxx/sessions", requestData);
        if (fail(response)) {
            String msg = String.format("Login %s error %d: %s",
                    ip, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        JSONObject respondData = response.getJSONObject("data");
        String token = respondData.getString("iBaseToken");
        String deviceId = respondData.getString("deviceid");

        request.setUrl(String.format("https://%s:%d/deviceManager/rest/%s", ip, port, deviceId));
        request.setHeaders("iBaseToken", token);
        this.request = request;
    }

    void logout() {
        try {
            request.delete("/sessions");
            request.close();
        } catch (Exception e) {
            // Ignore any exception here
        } finally {
            request = null;
        }
    }

    JSONObject createVolume(String name, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        requestData.put("PARENTID", poolId);
        requestData.put("CAPACITY", capacity / 512);

        if (allocType == ALLOC_TYPE.THIN) {
            requestData.put("ALLOCTYPE", 1);
        } else {
            requestData.put("ALLOCTYPE", 0);
        }

        JSONObject response = (JSONObject)request.post("/lun", requestData);
        if (fail(response)) {
            String msg = String.format("Create volume %s error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    void deleteVolume(String volumeId) throws Exception {
        JSONObject response = (JSONObject)request.delete(String.format("/lun/%s", volumeId));
        if (getErrorCode(response) == ERROR_CODE.VOLUME_NOT_EXIST.getValue()) {
            // Volume doesn't exist, return success
            return;
        }

        if (fail(response)) {
            String msg = String.format("Delete volume %s error %d: %s",
                    volumeId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }
    }

    JSONArray listVolumes(String poolId) throws Exception {
        String lunCountUrl;
        if (!poolId.isEmpty()) {
            lunCountUrl = String.format("/lun/count?filter=PARENTID::%s", poolId);
        } else {
            lunCountUrl = String.format("/lun/count");
        }

        JSONObject countResponse = (JSONObject)request.get(lunCountUrl);
        if (fail(countResponse)) {
            String msg = String.format("Get lun count error %d: %s",
                    getErrorCode(countResponse), getErrorDescription(countResponse));
            throw new Exception(msg);
        }

        JSONObject countData = countResponse.getJSONObject("data");
        int count = countData.getInt("COUNT");
        JSONArray luns = new JSONArray();

        for (int i = 0; i < count; i += 100) {
            String batchQueryLunUrl;

            if (!poolId.isEmpty()) {
                batchQueryLunUrl = String.format("/lun?filter=PARENTID::%s&range=[%d-%d]", poolId, i, i + 100);
            } else {
                batchQueryLunUrl = String.format("/lun?range=[%d-%d]", i, i + 100);
            }

            JSONObject lunsResponse = (JSONObject)request.get(batchQueryLunUrl);
            if (fail(lunsResponse)) {
                String msg = String.format("Batch get luns error %d: %s",
                        getErrorCode(lunsResponse), getErrorDescription(lunsResponse));
                throw new Exception(msg);
            }

            if (!lunsResponse.has("data")) {
                break;
            }

            JSONArray data = lunsResponse.getJSONArray("data");
            for (Object lun: data) {
                luns.put(lun);
            }
        }

        return luns;
    }

    JSONArray listStoragePools() throws Exception {
        JSONObject response = (JSONObject)request.get("/storagepool");
        if (fail(response)) {
            String msg = String.format("Get storage pools error %d: %s",
                    getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        return response.getJSONArray("data");
    }

    JSONObject getStoragePool(String poolId) throws Exception {
        JSONObject response = (JSONObject)request.get(String.format("/storagepool/%s", poolId));
        if (fail(response)) {
            String msg = String.format("Get storage pool %s error %d: %s",
                    poolId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    private JSONObject getInitiator(String iniType, String initiator) throws Exception {
        String encoded = URLEncoder.encode(initiator.replace(":", "\\:"), "utf-8");
        JSONObject response = (JSONObject)request.get(String.format("/%s?filter=ID::%s", iniType, encoded));
        if (fail(response)) {
            String msg = String.format("Get %s %s error %d: %s",
                    iniType, initiator, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        JSONArray data = response.getJSONArray("data");
        if (data.isEmpty()) {
            return null;
        }

        return data.getJSONObject(0);
    }

    JSONObject getISCSIInitiator(String initiator) throws Exception {
        return this.getInitiator("iscsi_initiator", initiator);
    }

    JSONObject getFCInitiator(String initiator) throws Exception {
        return this.getInitiator("fc_initiator", initiator);
    }

    JSONObject createHost(String name, HOST_OS_TYPE osType) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        switch (osType) {
            case LINUX:
                requestData.put("OPERATIONSYSTEM", 0);
                break;
            case WINDOWS:
                requestData.put("OPERATIONSYSTEM", 1);
                break;
            case AIX:
                requestData.put("OPERATIONSYSTEM", 4);
                break;
            case ESXI:
                requestData.put("OPERATIONSYSTEM", 7);
                break;
            default:
                requestData.put("OPERATIONSYSTEM", 0);
                break;
        }

        JSONObject response = (JSONObject)request.post("/host", requestData);
        if (fail(response)) {
            String msg = String.format("Create host %s error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    JSONObject getHostById(String id) throws Exception {
        JSONObject response = (JSONObject)request.get(String.format("/host/%s", id));
        if (fail(response)) {
            String msg = String.format("Get host by ID %s error %d: %s",
                    id, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    JSONArray getHostsByLun(String lunId) throws Exception {
        JSONObject response = (JSONObject)request.get(
                String.format("/host/associate?ASSOCIATEOBJTYPE=11&ASSOCIATEOBJID=%s", lunId));
        if (fail(response)) {
            String msg = String.format("Get hosts by LUN ID %s error %d: %s",
                    lunId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        return response.getJSONArray("data");
    }

    private void addInitiatorToHost(String type, String initiator, String hostId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("PARENTTYPE", 21);
        requestData.put("PARENTID", hostId);

        JSONObject response = (JSONObject)request.put(String.format("/%s/%s", type, initiator), requestData);
        if (fail(response)) {
            String msg = String.format("Add initiator %s of type %s to host error %d: %s",
                    initiator, type, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }
    }

    void addISCSIInitiatorToHost(String initiator, String hostId) throws Exception {
        addInitiatorToHost("iscsi_initiator", initiator, hostId);
    }

    void addFCInitiatorToHost(String initiator, String hostId) throws Exception {
        addInitiatorToHost("fc_initiator", initiator, hostId);
    }

    JSONArray getHostsByHostGroup(String hostGroupId) throws Exception {
        JSONObject response = (JSONObject)request.get(
                String.format("/host/associate?ASSOCIATEOBJTYPE=14&ASSOCIATEOBJID=%s", hostGroupId));
        if (fail(response)) {
            String msg = String.format("Batch query host in group %s error %d: %s",
                    hostGroupId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        return response.getJSONArray("data");
    }

    JSONObject createHostGroup(String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);

        JSONObject response = (JSONObject)request.post("/hostgroup", requestData);
        if (fail(response)) {
            String msg = String.format("Create hostgroup %s error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    void addHostToHostGroup(String hostId, String hostGroupId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", hostGroupId);
        requestData.put("ASSOCIATEOBJTYPE", 21);
        requestData.put("ASSOCIATEOBJID", hostId);

        JSONObject response = (JSONObject)request.post("/hostgroup/associate", requestData);
        if (fail(response)) {
            String msg = String.format("Associate host %s to hostgroup %s error %d: %s",
                    hostId,
                    hostGroupId,
                    getErrorCode(response),
                    getErrorDescription(response));
            throw new Exception(msg);
        }
    }

    JSONArray getHostGroupsByHost(String hostId) throws Exception {
        JSONObject response = (JSONObject)request.get(
                String.format("/hostgroup/associate?ASSOCIATEOBJTYPE=21&ASSOCIATEOBJID=%s", hostId));
        if (fail(response)) {
            String msg = String.format("Batch query hostgroup which host %s belongs to error %d: %s",
                    hostId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        return response.getJSONArray("data");
    }

    JSONObject createMappingView(String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);

        JSONObject response = (JSONObject)request.post("/mappingview", requestData);
        if (fail(response)) {
            String msg = String.format("Create mappingview %s error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    JSONArray getMappingViewsByHostGroup(String hostGroupId) throws Exception {
        JSONObject response = (JSONObject)request.get(
                String.format("/mappingview/associate?ASSOCIATEOBJTYPE=14&ASSOCIATEOBJID=%s", hostGroupId));
        if (fail(response)) {
            String msg = String.format("Batch query mappingview which hostgroup %s belongs to error %d: %s",
                    hostGroupId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        return response.getJSONArray("data");
    }

    JSONObject createLunGroup(String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        requestData.put("APPTYPE", 0);

        JSONObject response = (JSONObject)request.post("/lungroup", requestData);
        if (fail(response)) {
            String msg = String.format("Create lungroup %s error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    JSONObject getLunGroupByMappingView(String mappingViewId) throws Exception {
        JSONObject response = (JSONObject)request.get(
                String.format("/lungroup/associate?ASSOCIATEOBJTYPE=245&ASSOCIATEOBJID=%s", mappingViewId));
        if (fail(response)) {
            String msg = String.format("Batch query lungroup associated to mappingview %s error %d: %s",
                    mappingViewId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        JSONArray data = response.getJSONArray("data");
        if (data.isEmpty()) {
            return null;
        }

        return data.getJSONObject(0);
    }

    void addLunToLunGroup(String lunId, String lunGroupId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", lunGroupId);
        requestData.put("ASSOCIATEOBJTYPE", 11);
        requestData.put("ASSOCIATEOBJID", lunId);

        JSONObject response = (JSONObject)request.post("/lungroup/associate", requestData);
        if (fail(response)) {
            String msg = String.format("Add lun %s to lungroup %s error %d: %s",
                    lunId,
                    lunGroupId,
                    getErrorCode(response),
                    getErrorDescription(response));
            throw new Exception(msg);
        }
    }

    void removeLunFromLunGroup(String lunId, String lunGroupId) throws Exception {
        JSONObject response = (JSONObject)request.delete(
                String.format("/lungroup/associate?ID=%s&ASSOCIATEOBJTYPE=11&ASSOCIATEOBJID=%s", lunGroupId, lunId));
        if (fail(response)) {
            String msg = String.format("Remove lun %s from lungroup %s error %d: %s",
                    lunId,
                    lunGroupId,
                    getErrorCode(response),
                    getErrorDescription(response));
            throw new Exception(msg);
        }
    }

    void associateGroupToMappingView(String groupId, int groupType, String mappingViewId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", mappingViewId);
        requestData.put("ASSOCIATEOBJTYPE", groupType);
        requestData.put("ASSOCIATEOBJID", groupId);

        JSONObject response = (JSONObject)request.put("/mappingview/create_associate", requestData);
        if (fail(response)) {
            String msg = String.format("Associate group %s to mappingview %s error %d: %s",
                    groupId,
                    mappingViewId,
                    getErrorCode(response),
                    getErrorDescription(response));
            throw new Exception(msg);
        }
    }

    JSONArray getLunGroupsByLun(String volumeId) throws Exception {
        JSONObject response = (JSONObject)this.request.get(
                String.format("/lungroup/associate?ASSOCIATEOBJTYPE=11&ASSOCIATEOBJID=%s", volumeId));
        if (fail(response)) {
            String msg = String.format("Get lungroup which lun %s belongs to error %d: %s",
                    volumeId, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        return response.getJSONArray("data");
    }

    private JSONObject getObjectByName(String type, String name) throws Exception {
        JSONObject response = (JSONObject)request.get(String.format("/%s?filter=NAME::%s", type, name));
        if (fail(response)) {
            String msg = String.format("Get %s by name %s error %d: %s",
                    type, name, getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            return null;
        }

        JSONArray data = response.getJSONArray("data");
        if (data.isEmpty()) {
            return null;
        }

        return data.getJSONObject(0);
    }

    JSONObject getHostByName(String name) throws Exception {
        return getObjectByName("host", name);
    }

    JSONObject getHostGroupByName(String name) throws Exception {
        return getObjectByName("hostgroup", name);
    }

    JSONObject getLunGroupByName(String name) throws Exception {
        return getObjectByName("lungroup", name);
    }

    JSONObject getMappingViewByName(String name) throws Exception {
        return getObjectByName("mappingview", name);
    }

    JSONObject getSystem() throws Exception {
        JSONObject response = (JSONObject)request.get("/system/");
        if (fail(response)) {
            String msg = String.format("Get system info error %d: %s",
                    getErrorCode(response), getErrorDescription(response));
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }
}
