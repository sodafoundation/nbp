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
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONTokener;
import org.json.JSONArray;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.exceptions.NotAuthorizedException;
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
            if(body.isEmpty())
                return null;
            Object json = new JSONTokener(body).nextValue();
            if(json instanceof JSONArray)
                return new JSONArray(body);
            return new JSONObject(body);
        }
    }

    private Request request;
    private StorageMO storage;
    
    private int getErrorCode(JSONObject response) {
        return (int) response.getLong("code");
    }

    private String getErrorMessage(JSONObject response) {
        return response.getString("message");
    }

    private boolean isFailed(JSONObject response) throws Exception {
        if(!response.has("code"))
            return false;
        int errorCode = getErrorCode(response);
        if (errorCode == 401) {
            throw new NotAuthorizedException(getErrorMessage(response));
        }
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
		JSONObject requestData = new JSONObject();
		JSONObject userField = new JSONObject();
		JSONObject domainField = new JSONObject();
		JSONObject passwdField = new JSONObject();
		JSONObject identityField = new JSONObject();
		JSONArray passwdArray = new JSONArray();
		JSONObject authField = new JSONObject();
		JSONObject scopeField = new JSONObject();
		JSONObject projectField = new JSONObject();
		// create post body required for login
		domainField.put("id", OPENSDS_DOMAIN.getValue());
		userField.put("name", user);
		userField.put("password", password);
		userField.put("domain", domainField);
		passwdField.put("user", userField);
		passwdArray.put("password");
		identityField.put("methods", passwdArray);
		identityField.put("password", passwdField);
		authField.put("identity", identityField);
		projectField.put("name", OPENSDS_TENANT.getValue());
		projectField.put("domain", domainField);
		scopeField.put("project", projectField);
		authField.put("scope", scopeField);
		requestData.put("auth", authField);
		// prepare post request identity token service
		Request request = new Request(ip, port, new Handler());
		//set default request parameters
		request.setHeaders("Content-Type", "application/json");
		request.setUrl(String.format("http://%s/identity/v3", ip));

		JSONObject response = (JSONObject)request.post("/auth/tokens", requestData);
		if (isFailed(response)) {
			String msg = String.format("Login User %s error %d: %s",
					user, getErrorCode(response), getErrorMessage(response));
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
            // Ignore any exception here
        } finally {
            request = null;
        }
    }

    void findDeviceInfo(String ip, int port) throws Exception {
	    request.setUrl(String.format("http://%s:%d/v1beta", ip, port));
	    JSONObject response = (JSONObject)request.get(String.format("/"));
	    if (isFailed(response)) {
            String msg = String.format("OpenSDS Device Info error %d: %s",
                    getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
	    storage = new StorageMO(OPENSDS_STORAGENAME.getValue(), response.getString("name"),
	    "", response.getString("status"), OPENSDS_VENDOR.getValue());
    }

    StorageMO getDeviceInfo() {
        return storage;
    }

    JSONObject createVolume(String name, String description, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("name", name);
        requestData.put("size", capacity);
        requestData.put("description", description);
        String availabilityZone = OPENSDS_AVAILABILITYZONE.getValue();
		if(!availabilityZone.isEmpty())
			requestData.put("availabilityZone", availabilityZone);

        JSONObject response = (JSONObject)request.post("/block/volumes", requestData);

        if (isFailed(response)) {
            String msg = String.format("Create volume %s error %d: %s",
                    name, getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
        return response;
    }

    void deleteVolume(String volumeId) throws Exception {
        JSONObject response = (JSONObject)request.delete(String.format("/block/volumes/%s", volumeId));
        if (isFailed(response)) {
            String msg = String.format("Delete volume %s error %d: %s",
                    volumeId, getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
    }

    JSONArray listVolumes(String poolId) throws Exception {
        String volumeUrl;
        if (!poolId.isEmpty()) {
            volumeUrl = String.format("/block/volumes?poolId=%s", poolId);
        } else {
            volumeUrl = String.format("/block/volumes");
        }
        JSONArray response = (JSONArray)request.get(volumeUrl);
        if(response.isEmpty()) {
            String msg = String.format("List Volumes error: No Volumes Found");
            throw new Exception(msg);
        }

        return response;
    }

    JSONArray listVolumes(String filterKey, String filterValue) throws Exception {
        String volumeUrl;
        volumeUrl = String.format("/block/volumes?%s=%s", filterKey, filterValue);
        JSONArray response = (JSONArray)request.get(volumeUrl);
        if(response.isEmpty()) {
            String msg = String.format("List Volumes error: No Volumes Found");
            throw new Exception(msg);
        }

        return response;
    }

    JSONArray listStoragePools() throws Exception {
        JSONArray response = (JSONArray)request.get("/pools");
        if(response.isEmpty()) {
             String msg = String.format("List Storage Pools error: No Pools Found");
               throw new Exception(msg);
        }

        return response;
    }

    JSONObject getStoragePool(String poolId) throws Exception {

        JSONObject response = (JSONObject)request.get(String.format("/pools/%s", poolId));
        if (isFailed(response)) {
            String msg = String.format("List Storage Pool %s error %d: %s",
                    poolId, getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
        return response;
    }

    JSONObject attachVolume(String volumeId, String initiator, String initiatorIp) throws Exception {
        JSONObject requestData = new JSONObject();
        JSONObject hostInfo = new JSONObject();
        requestData.put("volumeId", volumeId);
        hostInfo.put("initiator", initiator);
        hostInfo.put("ip", initiatorIp);
        requestData.put("hostInfo", hostInfo);

        JSONObject response = (JSONObject)request.post("/block/attachments", requestData);

        if (isFailed(response)) {
            String msg = String.format("Attach volume %s error %d: %s",
                    volumeId, getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
        return response;
    }

    JSONObject detachVolume(String volumeId) throws Exception {
        JSONArray response = (JSONArray)request.get(String.format("/block/attachments?VolumeId=%s", volumeId));
        JSONObject attachment = (JSONObject)response.get(0);
        String attachmentId = (String) attachment.get("id");

        if(response.isEmpty() || attachmentId.isEmpty()) {
         String msg = String.format("Detach volume %s error: No Attachment Found",
                    volumeId);
            throw new Exception(msg);
        }

        JSONObject responseDelete = (JSONObject)request.delete(String.format("/block/attachments/%s", attachmentId));
        if (isFailed(responseDelete)) {
            String msg = String.format("Detach volume %s error %d: %s",
                    volumeId, getErrorCode(responseDelete), getErrorMessage(responseDelete));
            throw new Exception(msg);
        }

        JSONObject requestData = new JSONObject();
        requestData.put("status", VOLUME_STATUS.AVAILABLE.getValue());
        JSONObject responsePut = (JSONObject)request.put(String.format("/block/volumes/%s", volumeId), requestData);

        if (isFailed(responsePut)) {
            String msg = String.format("Detach volume %s error %d: %s",
                    volumeId, getErrorCode(responsePut), getErrorMessage(responsePut));
            throw new Exception(msg);
        }
        return responseDelete;
    }

    JSONObject getVolume(String volumeId) throws Exception {

        JSONArray response = (JSONArray)request.get(String.format("/block/volumes?wwn=%s", volumeId));
        if (response.isEmpty()) {
            String msg = String.format("List Volume for WWN error: No Volumes Found");
            throw new Exception(msg);
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
        JSONArray response = (JSONArray)request.get(snapshotUrl);
        if(response.isEmpty()) {
            String msg = String.format("List Volume Snapshots error: No Volumes Found");
            throw new Exception(msg);
        }

        return response;
    }

    JSONObject createVolumeSnapshot(String volumeId, String name) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("name", name);
        requestData.put("volumeId", volumeId);

        JSONObject response = (JSONObject)request.post("/block/snapshots", requestData);

        if (isFailed(response)) {
            String msg = String.format("Create volume snapshot %s error %d: %s",
                    name, getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
        return response;
    }

    void deleteVolumeSnapshot(String snapshotId) throws Exception {
        request.delete(String.format("/block/snapshots/%s", snapshotId));
    }

    void expandVolume(String volumeId, long capacity) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("newSize", capacity);

        JSONObject response = (JSONObject)request.post(String.format("/block/volumes/%s/resize", volumeId), requestData);

        if (isFailed(response)) {
            String msg = String.format("Expand Volume %s error %d: %s",
                    volumeId, getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
    }
}