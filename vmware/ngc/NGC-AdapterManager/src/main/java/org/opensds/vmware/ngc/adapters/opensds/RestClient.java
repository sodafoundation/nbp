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

import org.apache.http.Header;
import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.entity.StringEntity;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONTokener;
import org.json.JSONArray;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.models.ALLOC_TYPE;
import org.opensds.vmware.ngc.models.StorageMO;

import static org.opensds.vmware.ngc.adapters.opensds.Constants.*;


class RestClient {
    class Handler implements Request.RequestHandler {
        @Override
        public void setRequestBody(HttpEntityEnclosingRequestBase req, Object body) {
            StringEntity entity = new StringEntity(body.toString(), "utf-8");
            req.setEntity(entity);
        }

        @Override
        public Object parseResponseBody(String body) throws JSONException {
            Object json = new JSONTokener(body).nextValue();
            if(json instanceof JSONArray)
                return new JSONArray(body);
            return new JSONObject(body);
        }
    }

    private Request request;
    private StorageMO storage;
    private static final Log logger = LogFactory.getLog(RestClient.class);
    
    private int getErrorCode(JSONObject response) {
        return (int) response.getLong("code");
    }

    private String getErrorMessage(JSONObject response) {
        return response.getString("message");
    }

    private boolean isFailed(JSONObject response) throws Exception {
        if(!response.has("code") || response.isNull("code"))
            return false;
        return true;
    }

    private String getAuthToken(Header[] headers) 
    {
    	for (Header header : headers) {
    		if(header.getName().equals("X-Subject-Token"))
    		{
    			return header.getValue();
    		}
     	}
    	return null;
    }

	void login(String ip, int port, String user, String password) throws Exception {

		JSONObject domainField = new JSONObject();
		domainField.put("id", OPENSDS_DOMAIN.getValue());

		JSONObject userField = new JSONObject();
		userField.put("name", user);
		userField.put("password", password);
		userField.put("domain", domainField);

		JSONObject passwdField = new JSONObject();
		passwdField.put("user", userField);

		JSONArray passwdArray = new JSONArray();
		passwdArray.put("password");

		JSONObject identityField = new JSONObject();
		identityField.put("methods", passwdArray);
		identityField.put("password", passwdField);

		JSONObject projectField = new JSONObject();
		projectField.put("name", OPENSDS_TENANT.getValue());
		projectField.put("domain", domainField);

		JSONObject scopeField = new JSONObject();
		scopeField.put("project", projectField);

		JSONObject authField = new JSONObject();
		authField.put("identity", identityField);
		authField.put("scope", scopeField);

		// create post body required for login
		JSONObject requestData = new JSONObject();
		requestData.put("auth", authField);

		// prepare post request identity token service
		Request request = new Request(ip, port, new Handler());

		//set default request parameters
		request.setHeaders("Content-Type", "application/json");
		request.setUrl(String.format("http://%s/identity/v3", ip));

		logger.info("OpenSDS Storage Device Login Getting Auth Token");
		JSONObject response = (JSONObject)request.post("/auth/tokens", requestData);
		logger.debug(String.format("OpenSDS Storage Device Login Response: %s", response));

		if (isFailed(response)) {
			String msg = String.format("Login User %s Error %d: %s",
					user, getErrorCode(response), getErrorMessage(response));
			logger.error(String.format("OpenSDS Login Error: %s", msg));
			throw new Exception(msg);
		}
		// get Headers and select Auth Token
		Header[] headers = request.getResponseHeaders();
		String token = getAuthToken(headers);

		String tenantId = response.getJSONObject("token").getJSONObject("project").getString("id");
		request.setHeaders("X-Auth-Token", token);
		this.request = request;
		findDeviceInfo(ip, port);
		
		this.request.setUrl(String.format("http://%s:%d/%s/%s", ip, port, storage.model, tenantId));
	}

    void logout() {
        try {
             request.close();
        } catch (Exception e) {
             logger.error(String.format("OpenSDS Logout Error: %s", e));
        } finally {
             request = null;
        }
    }

    void findDeviceInfo(String ip, int port) throws Exception {
        logger.info(String.format("OpenSDS Getting Storage Device info for OpenSDS Storage Device "
                    + "with IP %s", ip));

        request.setUrl(String.format("http://%s:%d/v1beta", ip, port));
	    JSONObject response = (JSONObject)request.get(String.format("/"));
	    logger.debug(String.format("OpenSDS Storage Device Info Response: %s", response));

	    if (isFailed(response)) {
            String msg = String.format("OpenSDS Device Info Error %d: %s",
                    getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }

	    storage = new StorageMO(OPENSDS_STORAGENAME.getValue(), response.getString("name"),
	    "", response.getString("status"), OPENSDS_VENDOR.getValue());

	    logger.info(String.format("OpenSDS Storage Device: %s", storage));
    }

    StorageMO getDeviceInfo() {
        logger.info("OpenSDS Getting Storage Device Info");
        logger.info(String.format("OpenSDS Storage Device: %s", storage));
        return storage;
    }

    JSONObject createVolume(String name, String description, ALLOC_TYPE allocType, long capacity,
            String poolId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("name", name);
        requestData.put("size", capacity);
        requestData.put("description", description);
        String availabilityZone = OPENSDS_AVAILABILITYZONE.getValue();
		if(!availabilityZone.isEmpty())
			requestData.put("availabilityZone", availabilityZone);

		logger.info("----------OpenSDS Creating Volume----------");

        JSONObject response = (JSONObject)request.post("/block/volumes", requestData);
        logger.debug(String.format("OpenSDS Create Volume Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("Create volume %s Error %d: %s",
                    name, getErrorCode(response), getErrorMessage(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info("OpenSDS Volume Created.");
        return response;
    }

    void deleteVolume(String volumeId) throws Exception {
        logger.info(String.format("----------OpenSDS Deleting Volume for VolumeId "
                    + "%s----------", volumeId));

        JSONObject response = (JSONObject)request.delete(String.format("/block/volumes/%s",
                    volumeId));
        logger.debug(String.format("OpenSDS Delete Volume Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("Delete volume %s Error %d: %s",
                    volumeId, getErrorCode(response), getErrorMessage(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OpenSDS Volume for VolumeId %s Deleted. ", volumeId));
    }

    JSONArray listVolumes(String poolId) throws Exception {
        String volumeUrl;
        if (!poolId.isEmpty()) {
            volumeUrl = String.format("/block/volumes?PoolId=%s", poolId);
        } else {
            volumeUrl = String.format("/block/volumes");
        }

        logger.info("----------OpenSDS Listing Volumes----------");

        JSONArray response = (JSONArray)request.get(volumeUrl);
        logger.debug(String.format("OpenSDS List Volumes Response: %s", response));

        if(response.isEmpty()) {
            String msg = String.format("No Volumes Found");
            logger.error(String.format("OpenSDS List Volumes: %s", msg));
        }

        return response;
    }

    JSONArray listVolumes(String filterKey, String filterValue) throws Exception {
        logger.info("----------OpenSDS Listing Volumes----------");

        String volumeUrl;
        volumeUrl = String.format("/block/volumes?%s=%s", filterKey, filterValue);
        JSONArray response = (JSONArray)request.get(volumeUrl);
        logger.debug(String.format("OpenSDS List Volumes for the filter %s Response: %s",
                    filterKey, response));

        if(response.isEmpty()) {
            String msg = String.format("No Volumes Found");
            logger.error(String.format("OpenSDS List Volumes for the filter Error: %s %s",
                    filterKey, msg));
        }

        return response;
    }

    JSONArray listStoragePools() throws Exception {
        logger.info("----------OpenSDS Listing Storage Pools----------");

        JSONArray response = (JSONArray)request.get("/pools");
        logger.debug(String.format("OpenSDS List Storage Pools Response: %s", response));

        if(response.isEmpty()) {
             String msg = String.format("No Pools Found");
             logger.error(String.format("OpenSDS List Storage Pools: %s", msg));
        }

        return response;
    }

    JSONObject getStoragePool(String poolId) throws Exception {
		logger.info(String.format("----------OpenSDS Getting Info for Storage Pool %s----------",
                    poolId));

		JSONObject response = (JSONObject)request.get(String.format("/pools/%s", poolId));
        logger.debug(String.format("OpenSDS Getting Storage Pool for %s Response: %s", poolId,
                    response));

        if (isFailed(response)) {
            String msg = String.format("Get Storage Pool %s Error %d: %s",
                    poolId, getErrorCode(response), getErrorMessage(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        return response;
    }

    JSONObject attachVolume(String volumeId, String initiator, String initiatorIp) throws Exception {
        JSONObject hostInfo = new JSONObject();
        hostInfo.put("initiator", initiator);
        hostInfo.put("ip", initiatorIp);

        JSONObject requestData = new JSONObject();
        requestData.put("volumeId", volumeId);
        requestData.put("hostInfo", hostInfo);

        logger.info(String.format("----------OpenSDS Creating Volume attachment for Volume %s "
                    + "to %s----------", volumeId, initiatorIp));
        JSONObject response = (JSONObject)request.post("/block/attachments", requestData);
        logger.debug(String.format("OpenSDS Create Volume Attachment Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("Attach Volume %s Error %d: %s",
                    volumeId, getErrorCode(response), getErrorMessage(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info("OpenSDS Volume Attachment Created.");
        return response;
    }

    JSONObject detachVolume(String volumeId) throws Exception {
        logger.info(String.format("OpenSDS Getting Volume for VolumeId %s", volumeId));

        JSONArray response = (JSONArray)request.get(String.format("/block/attachments?VolumeId=%s",
                    volumeId));
        logger.debug(String.format("OpenSDS Get Volume Attachment for VolumeId %s Response: %s",
                    volumeId, response));

        JSONObject attachment = (JSONObject)response.get(0);
        String attachmentId = (String) attachment.get("id");

        if(response.isEmpty() || attachmentId.isEmpty()) {
            String msg = String.format("Detach Volume %s Error: No Attachment Found",
                    volumeId);
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("----------OpenSDS Deleting Volume Attachment for AttachmentId"
                    + " %s----------", attachmentId));
        JSONObject responseDelete = (JSONObject)request.delete(String.format("/block/attachments/%s",
                    attachmentId));
        logger.debug(String.format("OpenSDS DeleteVolumeAttachment for the attachmentId %s  Response: %s",
                    attachmentId, responseDelete));

        if (isFailed(responseDelete)) {
            String msg = String.format("Detach volume %s error %d: %s",
                    volumeId, getErrorCode(responseDelete), getErrorMessage(responseDelete));
            logger.error(msg);
            throw new Exception(msg);
        }

        JSONObject requestData = new JSONObject();
        requestData.put("status", VOLUME_STATUS.AVAILABLE.getValue());

        logger.info("OpenSDS Updating Volume Status to Available");
        JSONObject responsePut = (JSONObject)request.put(String.format("/block/volumes/%s", volumeId),
                    requestData);
        logger.debug(String.format("OpenSDS Update Volume for VolumeId %s Response: %s", volumeId,
                    responsePut));

        if (isFailed(responsePut)) {
            String msg = String.format("Detach volume %s Error %d: %s",
                    volumeId, getErrorCode(responsePut), getErrorMessage(responsePut));
            logger.error(msg);
            throw new Exception(msg);
        }

        return responseDelete;
    }

    JSONObject getVolume(String identifier) throws Exception {
        logger.info(String.format("----------OpenSDS Getting Volume for Identifier %s----------",
                    identifier));

        JSONArray response = (JSONArray)request.get(String.format("/block/volumes?wwn=%s", identifier));
        logger.debug(String.format("OpenSDS Getting Volume for Identifier %s Response: %s", identifier,
                    response));

        if (response.isEmpty()) {
            String msg = String.format("No Volumes Found");
            logger.error(String.format("OpenSDS Get Volume for %s Error: %s", identifier, msg));
        }

        JSONObject volume = (JSONObject) response.get(0);
        return volume;
    }

    JSONArray listVolumeSnapshot(String volumeId) throws Exception {
        String snapshotUrl;
        if (!volumeId.isEmpty()) {
            snapshotUrl = String.format("/block/snapshots?VolumeId=%s", volumeId);
        } else {
            snapshotUrl = String.format("/block/snapshots");
        }

        logger.info(String.format("----------OpenSDS Listing Snapshots for VolumeId %s----------",
                    volumeId));

        JSONArray response = (JSONArray)request.get(snapshotUrl);
        logger.debug(String.format("OpenSDS List Volume Snapshots for VolumeId %s Response: %s",
                    volumeId, response));

        if(response.isEmpty()) {
            String msg = String.format("No Snapshots Found");
            logger.error(String.format("OpenSDS List Volume Snapshots: %s", msg));
        }

        return response;
    }

    JSONObject createVolumeSnapshot(String volumeId, String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("name", name);
        requestData.put("volumeId", volumeId);

        logger.info(String.format("----------OpenSDS Creating Snapshot for VolumeId %s----------",
                    volumeId));
        JSONObject response = (JSONObject)request.post("/block/snapshots", requestData);
        logger.debug(String.format("OpenSDS CreateVolumeSnapshots for the VolumeId %s Response: %s",
                    volumeId, response));

        if (isFailed(response)) {
            String msg = String.format("Create Volume Snapshot %s Error %d: %s",
                    name, getErrorCode(response), getErrorMessage(response));
            logger.error(msg);
            throw new Exception(msg);
        }
        return response;
    }

    void deleteVolumeSnapshot(String snapshotId) throws Exception {
        logger.info(String.format("----------OpenSDS Deleting Snapshot for SnapshotId %s----------",
                    snapshotId));

        JSONObject response = (JSONObject)request.delete(String.format("/block/snapshots/%s",
                    snapshotId));
        logger.debug(String.format("OpenSDS Delete Volume Snapshot for the SnapshotId %s Response: %s",
                    snapshotId, response));
    }

    void expandVolume(String volumeId, long capacity) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("newSize", capacity);

        logger.info(String.format("----------OpenSDS Expanding Volume Size for VolumeId %s by"
                    + " %s----------", volumeId, capacity));
        
        JSONObject response = (JSONObject)request.post(String.format("/block/volumes/%s/resize",
                    volumeId), requestData);
        logger.debug(String.format("OpenSDS Expand Volume for the VolumeId %s Response: %s",
                    volumeId, response));

        if (isFailed(response)) {
            String msg = String.format("Expand Volume %s error %d: %s",
                    volumeId, getErrorCode(response), getErrorMessage(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info(String.format("OpenSDS Expanded Volume Size for VolumeId %s by %s",
                    volumeId, capacity));
    }
}