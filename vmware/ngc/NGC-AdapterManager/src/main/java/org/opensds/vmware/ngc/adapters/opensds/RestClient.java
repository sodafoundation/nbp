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

import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.entity.StringEntity;
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONTokener;
import org.json.JSONArray;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.exceptions.NotAuthorizedException;
import org.opensds.vmware.ngc.models.ALLOC_TYPE;

import org.openstack4j.api.OSClient.OSClientV3;
import org.openstack4j.model.common.Identifier;
import org.openstack4j.openstack.OSFactory;

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
    private OSClientV3 osClient;

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

    void login(String ip, int port, String user, String password) throws Exception {
        if(this.osClient != null) {
            logout();
        }

        if (this.request != null) {
            this.request.close();
        }

        Request request = new Request(ip, port, new Handler());
        request.setHeaders("Content-Type", "application/json");

        String endPoint = String.format("http://%s:80/identity/v3", ip);
        OSClientV3 osClient = OSFactory.builderV3()
                .endpoint(endPoint)
                .credentials(user, password, Identifier.byId("default"))
                .scopeToProject(Identifier.byName(user), Identifier.byName("Default"))
                .authenticate();
        String tenantId = "adminTenant";
        request.setUrl(String.format("http://%s:%d/v1beta/%s", ip, port,tenantId));
        request.setHeaders("X-Auth-Token", osClient.getToken().getId());

        this.osClient = osClient;
        this.request = request;
    }

    void logout() {
        try {
            osClient.identity().tokens().delete(osClient.getToken().getId());
            request.delete("/identity/v3");
            request.close();
        } catch (Exception e) {
            // Ignore any exception here
        } finally {
            osClient = null;
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
}
