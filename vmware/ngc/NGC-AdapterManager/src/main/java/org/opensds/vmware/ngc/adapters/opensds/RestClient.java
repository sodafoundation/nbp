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
    
    private int getErrorCode(JSONObject response) {
        return (int) response.getLong("code");
    }

    private String getErrorMessage(JSONObject response) {
        return response.getString("message");
    }

    private boolean fail(JSONObject response) throws Exception {
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
		domainField.put("id", "default");
		userField.put("name", "admin");
		userField.put("password", password);
		userField.put("domain", domainField);
		passwdField.put("user", userField);
		passwdArray.put("password");
		identityField.put("methods", passwdArray);
		identityField.put("password", passwdField);
		authField.put("identity", identityField);
		projectField.put("name", "admin");
		projectField.put("domain", domainField);
		scopeField.put("project", projectField);
		authField.put("scope", scopeField);
		requestData.put("auth", authField);
		// prepare post request identity token service
		Request request = new Request(ip, port, new Handler());
		request.setHeaders("Content-Type", "application/json");
		this.request = request;
		request.setUrl(String.format("http://%s/identity/v3", ip));
		request.post("/auth/tokens", requestData);
		// get Headers and select Auth Token
		Header[] headers = request.getResponseHeaders();
		String token = getAuthToken(headers);
		//set default request parameters
		request.setHeaders("Content-Type", "application/json");
		String tenantId = "adminTenant";
		request.setUrl(String.format("http://%s:%d/v1beta/%s", ip, port, tenantId));
		request.setHeaders("X-Auth-Token", token);
		this.request = request;
		
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

    JSONObject createVolume(String name, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("name", name);
        requestData.put("Description", "Test Volume Creation");
        requestData.put("size", capacity);
        requestData.put("AvailabilityZone", "default");

        JSONObject response = (JSONObject)request.post("/block/volumes", requestData);

        if (fail(response)) {
            String msg = String.format("Create volume %s error %d: %s",
                    name, getErrorCode(response), getErrorMessage(response));
            throw new Exception(msg);
        }
        return response;
    }

    void deleteVolume(String volumeId) throws Exception {
        JSONObject response = (JSONObject)request.delete(String.format("/block/volumes/%s", volumeId));
        if (fail(response)) {
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
        if (response.isEmpty()) {
            return null;
        }

        return response;
    }

    JSONArray listStoragePools() throws Exception {
        JSONArray response = (JSONArray)request.get("/pools");
        if (response.isEmpty()) {
            return null;
        }

        return response;
    }

    JSONObject getStoragePool(String poolId) throws Exception {

        return (JSONObject)request.get(String.format("/pools/%s", poolId));
    }
    public static void main( String[] args ) throws Exception
    {
    	RestClient client= new RestClient();
    	client.login("192.168.20.159", 50040, "admin", "opensds@123");
    	
    }
}
