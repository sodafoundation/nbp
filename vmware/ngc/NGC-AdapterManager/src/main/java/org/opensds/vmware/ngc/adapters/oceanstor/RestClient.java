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
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONArray;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.exceptions.NotAuthorizedException;
import org.opensds.vmware.ngc.models.ALLOC_TYPE;
import org.opensds.vmware.ngc.models.HOST_OS_TYPE;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.net.URLEncoder;
import java.util.ArrayList;
import java.util.List;

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
    private String ip;
    private int port;
    private String user;
    private String password;
    private static final Log logger = LogFactory.getLog(RestClient.class);

    RestClient(String ip, int port, String user, String password) {
        this.ip = ip;
        this.port = port;
        this.user = user;
        this.password = password;
    }

    private long getErrorCode(JSONObject response) {
        JSONObject error = response.getJSONObject("error");
        return error.getLong("code");
    }

    private String getErrorDescription(JSONObject response) {
        JSONObject error = response.getJSONObject("error");
        return error.getString("description");
    }

    private boolean isFailed(JSONObject response) throws Exception {
        long errorCode = getErrorCode(response);
        if (errorCode == -401) {
            throw new NotAuthorizedException(getErrorDescription(response));
        }

        return errorCode != 0;
    }

    void login() throws Exception {
        if (this.request != null) {
            this.request.close();
        }

        Request request = new Request(ip, port, new Handler());
        request.setHeaders("Content-Type", "application/json");

        JSONObject requestData = new JSONObject();
        requestData.put("username", user);
        requestData.put("password", password);
        requestData.put("scope", "0");

        logger.info("----------OceanStor Storage Device Login Getting Auth Token and "
                    + "Device Info----------");
        JSONObject response = (JSONObject)request.post(
                "/deviceManager/rest/xxxxx/sessions", requestData);
        logger.debug(String.format("OceanStor Storage Device Login Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("Login %s error %d: %s",
                    ip, getErrorCode(response), getErrorDescription(response));
            logger.error(String.format("OceanStor Logout Error: %s", msg));
            throw new Exception(msg);
        }

        JSONObject responseData = response.getJSONObject("data");
        String token = responseData.getString("iBaseToken");
        String deviceId = responseData.getString("deviceid");

        request.setUrl(String.format("https://%s:%d/deviceManager/rest/%s", ip, port, deviceId));
        request.setHeaders("iBaseToken", token);
        this.request = request;
    }

    void logout() {
        try {
            request.delete("/sessions");
            request.close();
        } catch (Exception e) {
            logger.error(String.format("OceanStor Logout Error: %s", e));
        } finally {
            request = null;
        }
    }

    JSONObject createVolume(String name, ALLOC_TYPE allocType, Long capacity, String poolId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        requestData.put("PARENTID", poolId);
        requestData.put("CAPACITY", capacity.longValue() / 512);

        if (allocType == ALLOC_TYPE.THIN) {
            requestData.put("ALLOCTYPE", 1);
        } else {
            requestData.put("ALLOCTYPE", 0);
        }

        logger.info("----------OceanStor Creating Volume----------");

        JSONObject response = (JSONObject)request.post("/lun", requestData);
        logger.debug(String.format("OceanStor Create Volume Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("OeanStor Create Volume %s Error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info("OceanStor Volume Created.");
        return response.getJSONObject("data");
    }

    void deleteVolume(String volumeId) throws Exception {
        logger.info(String.format("----------OceanStor Deleting Volume for VolumeId "
                    + "%s----------", volumeId));

        JSONObject response = (JSONObject)request.delete(String.format("/lun/%s", volumeId));
        logger.debug(String.format("OceanStor Delete Volume Response: %s", response));

        if (getErrorCode(response) == ERROR_CODE.VOLUME_NOT_EXIST.getValue()) {
            String msg = String.format("No Volumes Found");
            logger.error(String.format("OceanStor List Volumes: %s", msg));
            return;
        }

        if (isFailed(response)) {
            String msg = String.format("OceanStor Delete Volume %s Error %d: %s",
                    volumeId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Volume for VolumeId %s Deleted.", volumeId));
    }

    JSONArray listVolumes(String poolId) throws Exception {
        try {
            String lunCountUrl;
            if (!poolId.isEmpty()) {
                lunCountUrl = String.format("/lun/count?filter=PARENTID::%s", poolId);
            } else {
                lunCountUrl = String.format("/lun/count");
            }

            logger.info("----------OceanStor Listing Volumes----------");

            JSONObject countResponse = (JSONObject)request.get(lunCountUrl);
            logger.debug(String.format("OceanStor Lun Count Response: %s", countResponse));

            if (isFailed(countResponse)) {
                String msg = String.format("OceanStor Get Lun Count Error %d: %s",
                        getErrorCode(countResponse), getErrorDescription(countResponse));
                logger.error(msg);
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
                logger.debug(String.format("OceanStor List Lun Response: %s", countResponse));

                if (isFailed(lunsResponse)) {
                    String msg = String.format("Batch Get Luns Error %d: %s",
                            getErrorCode(lunsResponse), getErrorDescription(lunsResponse));
                    logger.error(msg);
                    throw new Exception(msg);
                }

                if (!lunsResponse.has("data")) {
                    String msg = String.format("No Data Found");
                    logger.error(String.format("OceanStor Listing Luns: %s", msg));
                    break;
                }

                for (Object lun: lunsResponse.getJSONArray("data")) {
                    luns.put(lun);
                }
            }

            return luns;
        }
        catch(Exception e) {
            logger.error(String.format("Error in Listing Volumes, Error Message is: %s", e));
            throw new JSONException("Error in Listing Volumes", e);
        }
    }

    JSONArray listVolumes(String filterKey, String filterValue) throws Exception {
       try {
            String lunCountUrl ="";
            if (!filterValue.isEmpty()) {
                lunCountUrl = String.format("/lun/count?filter=%s::%s", filterKey, filterValue);
            }

            logger.info("----------OceanStor Listing Volumes----------");

            JSONObject countResponse = (JSONObject)request.get(lunCountUrl);
            logger.debug(String.format("OceanStor List Volumes Response: %s", countResponse));

            if (isFailed(countResponse)) {
                String msg = String.format("Get Lun Count Error %d: %s",
                        getErrorCode(countResponse), getErrorDescription(countResponse));
                logger.error(msg);
                throw new Exception(msg);
            }

            JSONObject countData = countResponse.getJSONObject("data");
            int count = countData.getInt("COUNT");
            JSONArray luns = new JSONArray();

            for (int i = 0; i < count; i += 100) {
                String batchQueryLunUrl="";

                if (!filterValue.isEmpty()) {
                    batchQueryLunUrl = String.format("/lun?filter=%s::%s&range=[%d-%d]", filterKey, filterValue, i, i + 100);
                }

                JSONObject lunsResponse = (JSONObject)request.get(batchQueryLunUrl);
                logger.debug(String.format("OceanStor List Lun Response: %s", countResponse));

                if (isFailed(lunsResponse)) {
                    String msg = String.format("Batch Get Luns Error %d: %s",
                            getErrorCode(lunsResponse), getErrorDescription(lunsResponse));
                    logger.error(msg);
                    throw new Exception(msg);
                }

                if (!lunsResponse.has("data")) {
                    String msg = String.format("No Data Found");
                    logger.error(String.format("OceanStor Listing Luns: %s", msg));
                    break;
                }

                for (Object lun: lunsResponse.getJSONArray("data")) {
                    luns.put(lun);
                }
            }

            return luns;
        }
        catch(Exception e) {
            logger.error(String.format("Error in Listing Volumes, Error Message is: %s", e));
            throw new JSONException("Error in Listing Volumes", e);
        }
    }

    JSONArray listStoragePools() throws Exception {
        logger.info("----------OceanStor Listing Storage Pools----------");

        JSONObject response = (JSONObject)request.get("/storagepool");
        logger.debug(String.format("OceanStor List Storage Pools Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("List Storage Pools Error %d: %s",
                    getErrorCode(response), getErrorDescription(response));
            logger.error( msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Listing Storage Pools: %s", msg));
            return new JSONArray();
        }

        return response.getJSONArray("data");
    }

    JSONObject getStoragePool(String poolId) throws Exception {
        logger.info(String.format("----------OceanStor Getting Info for Storage Pool %s----------",
                    poolId));

        JSONObject response = (JSONObject)request.get(String.format("/storagepool/%s", poolId));
        logger.debug(String.format("OceanStor Getting Storage Pool for %s Response: %s", poolId,
                    response));

        if (isFailed(response)) {
            String msg = String.format("Get Storage Pool %s Error %d: %s",
                    poolId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    private JSONObject getInitiator(String iniType, String initiator) throws Exception {
        try {
            logger.info(String.format("----------OceanStor Getting Info for Initiator Type %s "
                    + "for Initiator %s----------", iniType, initiator));

            String encoded = URLEncoder.encode(initiator.replace(":", "\\:"), "utf-8");
            JSONObject response = (JSONObject)request.get(String.format("/%s?filter=ID::%s", iniType, encoded));
            logger.info(String.format("OceanStor Getting Info for Initiator Type %s "
                    + "for Initiator %s Response: %s", iniType, initiator, response));

            if (isFailed(response)) {
                String msg = String.format("Get %s %s error %d: %s",
                        iniType, initiator, getErrorCode(response), getErrorDescription(response));
                logger.error(msg);
                throw new Exception(msg);
            }

            if (!response.has("data")) {
                String msg = String.format("No Data Found");
                logger.error(String.format("OceanStor Getting Initiator Info: %s", msg));
                return null;
            }

            JSONArray data = response.getJSONArray("data");
            if (data.isEmpty()) {
                String msg = String.format("No Initiator Found");
                logger.error(String.format("OceanStor Getting Initiator Info: %s", msg));
                return null;
            }

            return data.getJSONObject(0);
        }
        catch(Exception e) {
            logger.error(String.format("Error in Getting Initiator Info, Error Message is: %s", e));
            throw new JSONException("Error in Getting Initiator Info", e);
       }
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
        logger.info(String.format("----------OceanStor Creating Host %s for OSType %s----------",
                    name, osType));

        JSONObject response = (JSONObject)request.post("/host", requestData);
        logger.debug(String.format("OceanStor Creating Host %s for OSType %s Response: %s", name,
                    osType, response));

        if (isFailed(response)) {
            String msg = String.format("Create Host %s Error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Host %s for OSType %s is Created.", name, osType));
        return response.getJSONObject("data");
    }

    JSONObject getHostById(String id) throws Exception {
        logger.info(String.format("----------OceanStor Getting Host for Id %s----------", id));

        JSONObject response = (JSONObject)request.get(String.format("/host/%s", id));
        logger.debug(String.format("OceanStor Getting Host for Id %s, Response: %s", id, response));

        if (isFailed(response)) {
            String msg = String.format("Get Host by ID %s error %d: %s",
                    id, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    JSONArray getHostsByLun(String lunId) throws Exception {
        logger.info(String.format("----------OceanStor Getting Hosts for LunId %s----------", lunId));

        JSONObject response = (JSONObject)request.get(
                String.format("/host/associate?ASSOCIATEOBJTYPE=11&ASSOCIATEOBJID=%s", lunId));
        logger.debug(String.format("OceanStor Getting Hosts for LunId %s, Response: %s", lunId, response));

        if (isFailed(response)) {
            String msg = String.format("Get hosts by LUN ID %s error %d: %s",
                    lunId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Getting Hosts for the LunId %s: %s", lunId, msg));
            return null;
        }

        return response.getJSONArray("data");
    }

    private void addInitiatorToHost(String type, String initiator, String hostId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("PARENTTYPE", 21);
        requestData.put("PARENTID", hostId);
        logger.info(String.format("----------OceanStor Adding Initiator %s of Type %s to Host %s----------",
                    initiator, type, hostId));

        JSONObject response = (JSONObject)request.put(String.format("/%s/%s", type, initiator), requestData);
        logger.debug(String.format("OceanStor Adding Initiator %s of Type %s to Host %s, Response:",
                    initiator, type, hostId, response));

        if (isFailed(response)) {
            String msg = String.format("Add Initiator %s of Type %s to Host Error %d: %s",
                    initiator, type, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Initiator %s of Type %s is added to Host %s",
                    initiator, type, hostId));
    }

    void addISCSIInitiatorToHost(String initiator, String hostId) throws Exception {
        addInitiatorToHost("iscsi_initiator", initiator, hostId);
    }

    void addFCInitiatorToHost(String initiator, String hostId) throws Exception {
        addInitiatorToHost("fc_initiator", initiator, hostId);
    }

    JSONArray getHostsByHostGroup(String hostGroupId) throws Exception {
        logger.info(String.format("----------OceanStor Getting Host for HostGroup %s----------",
                    hostGroupId));

        JSONObject response = (JSONObject)request.get(
                String.format("/host/associate?ASSOCIATEOBJTYPE=14&ASSOCIATEOBJID=%s", hostGroupId));
        logger.debug(String.format("OceanStor Getting Hosts for HostGroup %s, Response: %s",
                    hostGroupId, response));

        if (isFailed(response)) {
            String msg = String.format("Batch query Host in HostGroup %s Error %d: %s",
                    hostGroupId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Getting Hosts for HostGroup %s: %s", hostGroupId, msg));
            return null;
        }

        return response.getJSONArray("data");
    }

    JSONObject createHostGroup(String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        logger.info(String.format("----------OceanStor Creating HostGroup %s----------", name));

        JSONObject response = (JSONObject)request.post("/hostgroup", requestData);
        logger.debug(String.format("OceanStor Creating HostGroup %s, response: %s", name, response));

        if (isFailed(response)) {
            String msg = String.format("Create HostGroup %s Error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor HostGroup %s is Created.", name));
        return response.getJSONObject("data");
    }

    void addHostToHostGroup(String hostId, String hostGroupId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", hostGroupId);
        requestData.put("ASSOCIATEOBJTYPE", 21);
        requestData.put("ASSOCIATEOBJID", hostId);
        logger.info(String.format("----------OceanStor Adding Host %s to HostGroup %s----------",
                    hostId, hostGroupId));

        JSONObject response = (JSONObject)request.post("/hostgroup/associate", requestData);
        logger.debug(String.format("OceanStor Adding Host %s to HostGroup %s, Response: %s",
                    hostId, hostGroupId, response));

        if (isFailed(response)) {
            String msg = String.format("Associate Host %s to HostGroup %s Error %d: %s",
                    hostId,
                    hostGroupId,
                    getErrorCode(response),
                    getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Host %s is added to HostGroup %s.", hostId, hostGroupId));
    }

    JSONArray getHostGroupsByHost(String hostId) throws Exception {
        logger.info(String.format("----------OceanStor Getting HostGroup for Host %s----------",
                    hostId));

        JSONObject response = (JSONObject)request.get(
                String.format("/hostgroup/associate?ASSOCIATEOBJTYPE=21&ASSOCIATEOBJID=%s", hostId));
        logger.debug(String.format("OceanStor Getting HostGroup for Host %s, Response: %s",
                    hostId, response));

        if (isFailed(response)) {
            String msg = String.format("Batch query HostGroup for the Host %s Error %d: %s",
                    hostId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Getting HostGroups for Host %s: %s", hostId, msg));
            return null;
        }

        return response.getJSONArray("data");
    }

    JSONObject createMappingView(String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        logger.info(String.format("----------OceanStor Creating Mapping View %s----------", name));

        JSONObject response = (JSONObject)request.post("/mappingview", requestData);
        logger.debug(String.format("OceanStor Creating Mapping View %s, Response: %s", name, response));

        if (isFailed(response)) {
            String msg = String.format("Create MappingView %s Error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Mapping View %s Created.", name));
        return response.getJSONObject("data");
    }

    JSONArray getMappingViewsByHostGroup(String hostGroupId) throws Exception {
        logger.info(String.format("----------OceanStor Getting MappingViews for HostGroup %s----------",
                    hostGroupId));

        JSONObject response = (JSONObject)request.get(
                String.format("/mappingview/associate?ASSOCIATEOBJTYPE=14&ASSOCIATEOBJID=%s", hostGroupId));
        logger.debug(String.format("OceanStor Getting MappingViews for HostGroup %s, Response: %s",
                    hostGroupId, response));

        if (isFailed(response)) {
            String msg = String.format("Batch query MappingView for HostGroup %s Error %d: %s",
                    hostGroupId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Getting  MappingView for HostGroup %s: %s", hostGroupId, msg));
            return null;
        }

        return response.getJSONArray("data");
    }

    JSONObject createLunGroup(String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        requestData.put("APPTYPE", 0);
        logger.info(String.format("----------OceanStor Creating LunGroup %s----------", name));

        JSONObject response = (JSONObject)request.post("/lungroup", requestData);
        logger.debug(String.format("OceanStor Creating LunGroup %s, Response: %s", name, response));

        if (isFailed(response)) {
            String msg = String.format("Create LunGroup %s Error %d: %s",
                    name, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor LunGroup %s Created.", name));
        return response.getJSONObject("data");
    }

    JSONObject getLunGroupByMappingView(String mappingViewId) throws Exception {
        logger.info(String.format("----------OceanStor Getting LunGroups for MappingView %s----------",
                    mappingViewId));

        JSONObject response = (JSONObject)request.get(
                String.format("/lungroup/associate?ASSOCIATEOBJTYPE=245&ASSOCIATEOBJID=%s", mappingViewId));
        logger.debug(String.format("OceanStor Getting LunGroups for MappingView %s, Response: %s",
                    mappingViewId, response));

        if (isFailed(response)) {
            String msg = String.format("Batch query LunGroup associated to MappingView %s Error %d: %s",
                    mappingViewId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Getting LunGroup for MappingView  %s: %s", mappingViewId, msg));
            return null;
        }

        JSONArray data = response.getJSONArray("data");
        if (data.isEmpty()) {
            String msg = String.format("No LunGroups Found");
            logger.error(String.format("OceanStor Getting LunGroups for MappingView %s: %s", mappingViewId, msg));
            return null;
        }

        return data.getJSONObject(0);
    }

    void addLunToLunGroup(String lunId, String lunGroupId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", lunGroupId);
        requestData.put("ASSOCIATEOBJTYPE", 11);
        requestData.put("ASSOCIATEOBJID", lunId);
 
        logger.info(String.format("----------OceanStor Adding Lun %s to the Lun Group %s----------",
                    lunId, lunGroupId));

        JSONObject response = (JSONObject)request.post("/lungroup/associate", requestData);
        logger.debug(String.format("OceanStor Add Lun %s to the Lun Groups %s Response: %s",
                    lunId, lunGroupId, response));

        if (isFailed(response)) {
            String msg = String.format("Add Lun %s to the LunGroup %s Error %d: %s",
                    lunId,
                    lunGroupId,
                    getErrorCode(response),
                    getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }
        logger.info(String.format("OceanStor Lun %s is added to the Lun Group %s.", lunId, lunGroupId));
    }

    void removeLunFromLunGroup(String lunId, String lunGroupId) throws Exception {
        logger.info(String.format("----------OceanStor Removing Lun %s from the Lun Group %s----------",
                    lunId, lunGroupId));

        JSONObject response = (JSONObject)request.delete(
                String.format("/lungroup/associate?ID=%s&ASSOCIATEOBJTYPE=11&ASSOCIATEOBJID=%s",
                    lunGroupId, lunId));
        logger.debug(String.format("OceanStor Remove Lun from the Lun Groups for lunId %s Response: %s",
                    lunId, response));

        if (isFailed(response)) {
            String msg = String.format("Remove Lun %s from LunGroup %s Error %d: %s",
                    lunId,
                    lunGroupId,
                    getErrorCode(response),
                    getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }
        logger.info(String.format("OceanStor Lun %s is removed from the Lun Group %s.", lunId, lunGroupId));
    }

    void associateGroupToMappingView(String groupId, Integer groupType, String mappingViewId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", mappingViewId);
        requestData.put("ASSOCIATEOBJTYPE", groupType.intValue());
        requestData.put("ASSOCIATEOBJID", groupId);

        logger.info(String.format("----------OceanStor Associating Lun Group %s to Mapping View %s----------",
                    groupId, mappingViewId));

        JSONObject response = (JSONObject)request.put("/mappingview/create_associate", requestData);
        logger.debug(String.format("OceanStor Associating Lun Group %s to Mapping View %s Response: %s",
                    groupId, mappingViewId, response));

        if (isFailed(response)) {
            String msg = String.format("Associate group %s to mappingview %s error %d: %s",
                    groupId,
                    mappingViewId,
                    getErrorCode(response),
                    getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Lun Group %s is associated to Mapping View %s",
                    groupId, mappingViewId));
    }

    JSONArray getLunGroupsByLun(String volumeId) throws Exception {
        logger.info(String.format("----------OceanStor Getting Lun Groups for volumeId %s----------",
                    volumeId));

        JSONObject response = (JSONObject)this.request.get(
                String.format("/lungroup/associate?ASSOCIATEOBJTYPE=11&ASSOCIATEOBJID=%s", volumeId));
        logger.debug(String.format("OceanStor Lun Group for VolumeId %s Response: %s",
                    volumeId, response));

        if (isFailed(response)) {
            String msg = String.format("Get Lun Groups for the lun %s Error %d: %s",
                    volumeId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Lun Groups for %s Error: %s", volumeId, msg));
            return null;
        }

        return response.getJSONArray("data");
    }

    private JSONObject getObjectByName(String type, String name) throws Exception {
        logger.info(String.format("----------OceanStor Getting Results with Type %s for %s----------",
                    type, name));

        JSONObject response = (JSONObject)request.get(String.format("/%s?filter=NAME::%s", type, name));
        logger.debug(String.format("OceanStor Getting Results with Type %s for %s, Result: %s",
                    type, name, response));

        if (isFailed(response)) {
            String msg = String.format("Get %s by Name %s Error %d: %s",
                    type, name, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        if (!response.has("data")) {
            String msg = String.format("No Data Found");
            logger.error(String.format("OceanStor Getting Result with Type %s for %s, Error: %s",
                    type, name, msg));
            return null;
        }

        JSONArray data = response.getJSONArray("data");
        if (data.isEmpty()) {
            String msg = String.format("No Results Found");
            logger.error(String.format("OceanStor Getting Result with Type %s for %s, Error: %s",
                    type, name, msg));
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
        logger.info("----------OceanStor Getting System Information----------");

        JSONObject response = (JSONObject)request.get("/system/");
        logger.debug(String.format("OceanStor Getting System Info Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("Get System Info Error %d: %s",
                    getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    JSONObject getVolumeById(String lunId) throws Exception {
        logger.info(String.format("----------OceanStor Getting Lun for LunId %s----------",
                    lunId));
    	
        JSONObject response = (JSONObject)request.get(String.format("/lun/%s", lunId));
        logger.debug(String.format("OceanStor Getting Lun for LunId %s Response: %s",
                    lunId, response));

        if (isFailed(response)) {
            String msg = String.format("Get Lun for LunId %s error %d: %s",
                    lunId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        return response.getJSONObject("data");
    }

    JSONArray listSnapshots(String volumeId) throws Exception {
        logger.info(String.format("----------OceanStor Listing Snapshots for VolumeId %s----------",
                    volumeId));

        JSONObject countResponse = (JSONObject)request.get(
                String.format("/snapshot/count?filter=PARENTID::%s", volumeId));
        logger.debug(String.format("OceanStor Lun Count for VolumeId %s Response: %s",
                    volumeId, countResponse));

        if (isFailed(countResponse)) {
            String msg = String.format("Get Lun Count for the lun %s Error %d: %s",
                    volumeId, getErrorCode(countResponse), getErrorDescription(countResponse));
            logger.error(msg);
            throw new Exception(msg);
        }

        JSONObject countData = countResponse.getJSONObject("data");
        JSONArray snapshots = new JSONArray();

        for (int i = 0; i < countData.getInt("COUNT"); i += 100) {
            JSONObject snapshotsResponse = (JSONObject)request.get(
                    String.format("/snapshot?filter=PARENTID::%s&range=[%d-%d]", volumeId, i, i + 100));

            logger.debug(String.format("OceanStor List Volume Snapshots for VolumeId %s Response: %s",
                    volumeId, snapshotsResponse));
            
            if (isFailed(snapshotsResponse)) {
                String msg = String.format("Batch Get Snapshots for the lun %s error %d: %s",
                        volumeId, getErrorCode(snapshotsResponse), getErrorDescription(snapshotsResponse));
                logger.error(msg);
                throw new Exception(msg);
            }

            if (!snapshotsResponse.has("data")) {
                break;
            }

            for (Object snapshot: snapshotsResponse.getJSONArray("data")) {
                snapshots.put(snapshot);
            }
        }

        return snapshots;
    }

    private JSONObject createLunSnapshot(String volumeId, String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("NAME", name);
        requestData.put("PARENTID", volumeId);

        logger.info(String.format("----------OceanStor Creating Snapshot for VolumeId %s----------",
                    volumeId));

        JSONObject response = (JSONObject)request.post("/snapshot", requestData);
        logger.debug(String.format("OceanStor CreateVolumeSnapshots for the VolumeId %s Response: %s",
                    volumeId, response));

        if (isFailed(response)) {
            String msg = String.format("Create Snapshot %s for %s Error %d: %s",
                    name, volumeId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Snapshot Created for VolumeId %s.", volumeId));
        return response.getJSONObject("data");
    }

    private void activateLunSnapshot(String snapshotId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("SNAPSHOTLIST", String.format("[%s]", snapshotId));

        logger.info(String.format("----------Oceanstor Activating Lun Snapshot for SnapshotId"
                    + " %s----------", snapshotId));

        JSONObject response = (JSONObject)request.post("/snapshot/activate", requestData);
        logger.debug(String.format("OceanStor Activate Lun Snapshot for the SnapshotId %s Response: %s",
                    snapshotId, response));

        if (isFailed(response)) {
            String msg = String.format("Activate Lun Snapshot %s Error %d: %s",
                    snapshotId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("Oceanstor Lun Snapshot Activated for SnapshotId %s.", snapshotId));
    }

    private void deactivateLunSnapshot(String snapshotId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", snapshotId);

        logger.info(String.format("----------Oceanstor Deactivating Lun Snapshot for SnapshotId"
                    + " %s----------", snapshotId));

        JSONObject response = (JSONObject)request.put("/snapshot/stop", requestData);
        logger.debug(String.format("OceanStor Deactivate Lun Snapshot for the SnapshotId %s Response: %s",
                    snapshotId, response));

        if (getErrorCode(response) == ERROR_CODE.SNAPSHOT_NOT_EXIST.getValue()) {
            logger.error(String.format("Deactivate Lun Snapshot Error: No Snapshot is Found"
                    + " with SnapshotId %s", snapshotId));
            return;
        }

        if (getErrorCode(response) == ERROR_CODE.SNAPSHOT_NOT_ACTIVATED.getValue()) {
            logger.error(String.format("Deactivate Lun Snapshot Error: Snapshot is not Activated"
                    + " with SnapshotId %s", snapshotId));
            return;
        }

        if (isFailed(response)) {
            String msg = String.format("Deactivate Lun Snapshot %s Error %d: %s",
                    snapshotId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("Oceanstor Lun Snapshot Deactivated for SnapshotId %s.", snapshotId));
    }

    private void deleteLunSnapshot(String snapshotId) throws Exception {
        logger.info(String.format("----------Oceanstor Deleting Lun Snapshot for SnapshotId"
                    + " %s----------", snapshotId));

    	JSONObject response = (JSONObject)request.delete(String.format("/snapshot/%s", snapshotId));
        logger.debug(String.format("OceanStor Delete Lun Snapshot for the SnapshotId %s Response: %s",
                    snapshotId, response));
    	
        if (getErrorCode(response) == ERROR_CODE.SNAPSHOT_NOT_EXIST.getValue()) {
            logger.error(String.format("Delete Lun Snapshot Error: No Snapshot is Found"
                    + " with SnapshotId %s", snapshotId));
            return;
        }

        if (isFailed(response)) {
            String msg = String.format("Delete Lun Snapshot %s Error %d: %s",
                    snapshotId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("Oceanstor Lun Snapshot Deleted for SnapshotId %s.", snapshotId));
    }

    void createVolumeSnapshot(String volumeId, String name) throws Exception {
        String snapshotId = "";

        try {
            JSONObject snapshot = createLunSnapshot(volumeId, name);
            logger.info(String.format("OceanStor Snapshot: %s", snapshot));

            snapshotId = snapshot.getString("ID");
            activateLunSnapshot(snapshotId);
        }catch (JSONException e) {
            logger.error(String.format("Oceanstor Create Volume Snapshot Error Message is: %s", e));
            throw new Exception("Error in Creating Volume Snapshot ", e);
        }catch (Exception e) {
            String msg = String.format("Activating Snapshot for VolumeId %s Failed, "
                    + "Deleting the Created Snapshot with SnapshotId %s", volumeId, snapshotId);
            logger.error(msg);
            throw new Exception(msg);
        }finally {
            if(!snapshotId.isEmpty()) {
                deleteLunSnapshot(snapshotId);
            }
        }

        logger.info(String.format("OceanStor Snapshot Created for VolumeId %s.", volumeId));
    }

    void deleteVolumeSnapshot(String snapshotId) throws Exception {
        deactivateLunSnapshot(snapshotId);
        deleteLunSnapshot(snapshotId);
    }

    void rollbackVolumeSnapshot(String snapshotId, String rollbackSpeed) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", snapshotId);
        requestData.put("ROLLBACKSPEED", rollbackSpeed);

        logger.info(String.format("----------OceanStor Rollingback Snapshot for snapshotId %s", snapshotId));

        JSONObject response = (JSONObject)request.put(String.format("/snapshot/rollback"), requestData);
        logger.debug(String.format("OceanStor Rollingback Snapshot for snapshotId %s, Response:",
                    snapshotId, response));

        if (isFailed(response)) {
            String msg = String.format("Rollback Snapshot %s Error %d: %s",
                    snapshotId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("Oceanstor Lun Snapshot Rollback Done for SnapshotId %s.", snapshotId));
    }
	
	void expandVolume(String volumeId, long capacity) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("ID", volumeId);
        requestData.put("CAPACITY", capacity / 512L);

        logger.info(String.format("----------OceanStor Expanding Volume Size for VolumeId %s by"
                    + " %s----------", volumeId, capacity));

        JSONObject response = (JSONObject)request.put(String.format("/lun/expand"), requestData);
        logger.debug(String.format("OceanStor Expand Volume for the VolumeId %s Response: %s",
                    volumeId, response));

        if (isFailed(response)) {
            String msg = String.format("Expand volume %s error %d: %s",
                    volumeId, getErrorCode(response), getErrorDescription(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OceanStor Expanded Volume Size for VolumeId %s by %s",
                    volumeId, capacity));
    }
}

class RestClientWrapper {
    private RestClient client;

    RestClientWrapper() {
        this.client = null;
    }

    void login(String ip, int port, String user, String password) throws Exception {
        this.client = new RestClient(ip, port, user, password);
        this.client.login();
    }

    void logout() {
        if (this.client != null) {
            this.client.logout();
        }
    }

    private void noReturnWrapper(String methodName, Object... parameters) throws Exception {
        if (this.client == null) {
            throw new Exception("Didn't login to any storage");
        }

        List<Class> parameterTypes = new ArrayList<>();
        for (Object object: parameters) {
            parameterTypes.add(object.getClass());
        }

        Method method = RestClient.class.getDeclaredMethod(methodName, parameterTypes.toArray(new Class[0]));

        try {
                method.invoke(this.client, parameters);
        } catch (Exception e) {
            if (e instanceof NotAuthorizedException) {
                this.client.login();
                method.invoke(this.client, parameters);
            } else {
                throw e;
            }
        }
    }

    private Object returnObjectWrapper(String methodName, Object... parameters) throws Exception {
        if (this.client == null) {
            throw new Exception("Didn't login to any storage");
        }

        List<Class> parameterTypes = new ArrayList<>();
        for (Object object: parameters) {
            parameterTypes.add(object.getClass());
        }

        Method method = RestClient.class.getDeclaredMethod(methodName, parameterTypes.toArray(new Class[0]));

        try {
            return method.invoke(this.client, parameters);
        } catch (InvocationTargetException e) {
            if (e.getTargetException() instanceof NotAuthorizedException) {
                this.client.login();
                return method.invoke(this.client, parameters);
            } else {
                throw (Exception)e.getTargetException();
            }
        }
    }

    JSONObject createVolume(String name, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception {
        return (JSONObject) returnObjectWrapper("createVolume", name, allocType, capacity, poolId);
    }

    void deleteVolume(String volumeId) throws Exception {
        noReturnWrapper("deleteVolume", volumeId);
    }

    JSONArray listVolumes(String poolId) throws Exception {
        return (JSONArray) returnObjectWrapper("listVolumes", poolId);
    }

    JSONArray listVolumes(String filterKey, String filterValue) throws Exception {
        return (JSONArray) returnObjectWrapper("listVolumes", filterKey, filterValue);
    }

    JSONArray listStoragePools() throws Exception {
        return (JSONArray) returnObjectWrapper("listStoragePools");
    }

    JSONObject getStoragePool(String poolId) throws Exception {
        return (JSONObject) returnObjectWrapper("getStoragePool", poolId);
    }

    JSONObject getISCSIInitiator(String initiator) throws Exception {
        return (JSONObject) returnObjectWrapper("getISCSIInitiator", initiator);
    }

    JSONObject getFCInitiator(String initiator) throws Exception {
        return (JSONObject) returnObjectWrapper("getFCInitiator", initiator);
    }

    JSONObject createHost(String name, HOST_OS_TYPE osType) throws Exception {
        return (JSONObject) returnObjectWrapper("createHost", name, osType);
    }

    JSONObject getHostById(String id) throws Exception {
        return (JSONObject) returnObjectWrapper("getHostById", id);
    }

    JSONArray getHostsByLun(String lunId) throws Exception {
        return (JSONArray) returnObjectWrapper("getHostsByLun", lunId);
    }

    void addISCSIInitiatorToHost(String initiator, String hostId) throws Exception {
        noReturnWrapper("addISCSIInitiatorToHost", initiator, hostId);
    }

    void addFCInitiatorToHost(String initiator, String hostId) throws Exception {
        noReturnWrapper("addFCInitiatorToHost", initiator, hostId);
    }

    JSONArray getHostsByHostGroup(String hostGroupId) throws Exception {
        return (JSONArray) returnObjectWrapper("getHostsByHostGroup", hostGroupId);
    }

    JSONObject createHostGroup(String name) throws Exception {
        return (JSONObject) returnObjectWrapper("createHostGroup", name);
    }

    void addHostToHostGroup(String hostId, String hostGroupId) throws Exception {
        noReturnWrapper("addHostToHostGroup", hostId, hostGroupId);
    }

    JSONArray getHostGroupsByHost(String hostId) throws Exception {
        return (JSONArray) returnObjectWrapper("getHostGroupsByHost", hostId);
    }

    JSONObject createMappingView(String name) throws Exception {
        return (JSONObject) returnObjectWrapper("createMappingView", name);
    }

    JSONArray getMappingViewsByHostGroup(String hostGroupId) throws Exception {
        return (JSONArray) returnObjectWrapper("getMappingViewsByHostGroup", hostGroupId);
    }

    JSONObject createLunGroup(String name) throws Exception {
        return (JSONObject) returnObjectWrapper("createLunGroup", name);
    }

    JSONObject getLunGroupByMappingView(String mappingViewId) throws Exception {
        return (JSONObject) returnObjectWrapper("getLunGroupByMappingView", mappingViewId);
    }

    void addLunToLunGroup(String lunId, String lunGroupId) throws Exception {
        noReturnWrapper("addLunToLunGroup", lunId, lunGroupId);
    }

    void removeLunFromLunGroup(String lunId, String lunGroupId) throws Exception {
        noReturnWrapper("removeLunFromLunGroup", lunId, lunGroupId);
    }

    void associateGroupToMappingView(String groupId, int groupType, String mappingViewId) throws Exception {
        noReturnWrapper("associateGroupToMappingView", groupId, groupType, mappingViewId);
    }

    JSONArray getLunGroupsByLun(String volumeId) throws Exception {
        return (JSONArray) returnObjectWrapper("getLunGroupsByLun", volumeId);
    }

    JSONObject getHostByName(String name) throws Exception {
        return (JSONObject) returnObjectWrapper("getHostByName", name);
    }

    JSONObject getHostGroupByName(String name) throws Exception {
        return (JSONObject) returnObjectWrapper("getHostGroupByName", name);
    }

    JSONObject getLunGroupByName(String name) throws Exception {
        return (JSONObject) returnObjectWrapper("getLunGroupByName", name);
    }

    JSONObject getMappingViewByName(String name) throws Exception {
        return (JSONObject) returnObjectWrapper("getMappingViewByName", name);
    }

    JSONObject getSystem() throws Exception {
        return (JSONObject) returnObjectWrapper("getSystem");
    }

    JSONObject getVolumeById(String volumeId) throws Exception {
        return (JSONObject) returnObjectWrapper("getVolumeById", volumeId);
    }

    JSONArray listSnapshots(String volumeId) throws Exception {
        return (JSONArray) returnObjectWrapper("listSnapshots", volumeId);
    }

    void createVolumeSnapshot(String volumeId, String name) throws Exception {
        noReturnWrapper("createVolumeSnapshot", volumeId, name);
    }

    void deleteVolumeSnapshot(String snapshotId) throws Exception {
        noReturnWrapper("deleteVolumeSnapshot", snapshotId);
    }

    void rollbackVolumeSnapshot(String snapshotId, String rollbackSpeed) throws Exception {
        noReturnWrapper("rollbackVolumeSnapshot", snapshotId, rollbackSpeed);
    }
	
	void expandVolume(String volumeId, long capacity) throws Exception {
        noReturnWrapper("expandVolume", volumeId, capacity);
    }
}
