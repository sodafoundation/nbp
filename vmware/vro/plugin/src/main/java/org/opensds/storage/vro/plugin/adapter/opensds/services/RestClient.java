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

package org.opensds.storage.vro.plugin.adapter.opensds.services;

import org.apache.http.Header;
import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.entity.StringEntity;
import org.apache.log4j.Logger;
import org.json.JSONException;
import org.json.JSONObject;
import org.json.JSONTokener;
import org.json.JSONArray;
import org.opensds.storage.vro.plugin.adapter.opensds.services.exceptions.NotAuthorizedException;
import org.opensds.storage.vro.plugin.adapter.opensds.model.StorageMO;
import org.opensds.storage.vro.plugin.adapter.opensds.model.ConnectMO;

class RestClient {
	private static final Logger logger = Logger.getLogger(RestClient.class);
	private static final String OPENSDS_HOST_ACCESSMODE="agentless";
	private static final String OPENSDS_AVAILABILITYZONE = "default";
	class Handler implements Request.RequestHandler {
		@Override
		public void setRequestBody(HttpEntityEnclosingRequestBase req, Object body) {
			StringEntity entity = new StringEntity(body.toString(), "utf-8");
			req.setEntity(entity);
		}

		@Override
		public Object parseResponseBody(String body) throws JSONException {
			if (body.isEmpty())
				return null;
			Object json = new JSONTokener(body).nextValue();
			if (json instanceof JSONArray)
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
		if (!response.has("code"))
			return false;
		int errorCode = getErrorCode(response);
		if (errorCode == 401) {
			throw new NotAuthorizedException(getErrorMessage(response));
		}
		return true;
	}

	private String getAuthToken(Header[] headers) {
		for (Header header : headers) {
			if (header.getName().equals("X-Subject-Token")) {
				return header.getValue();
			}
		}
		return null;
	}

	void login(OpenSDSInfo openSDSInfo) throws Exception {
		String ip = openSDSInfo.getHostName();
		int port = Integer.parseInt(openSDSInfo.getPort());
		String user = openSDSInfo.getUsername();
		String password = openSDSInfo.getPassword();

		if (!openSDSInfo.getauthEnabled()) {
			Request request = new Request(ip, port, new Handler());
			// set default request parameters
			request.setHeaders("Content-Type", "application/json");
			this.request = request;
			findDeviceInfo(ip, port);
			this.request.setUrl(String.format("http://%s:%d/%s/%s", ip, port, "v1beta", "admin"));
		} else {
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
			userField.put("name", user);
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
			// set default request parameters
			request.setHeaders("Content-Type", "application/json");
			request.setUrl(String.format("http://%s/identity/v3", ip));
			JSONObject response = (JSONObject) request.post("/auth/tokens", requestData);
			if (isFailed(response)) {
				String msg = String.format("Login User %s error %d: %s", user, getErrorCode(response),
						getErrorMessage(response));
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
		JSONObject response = (JSONObject) request.get(String.format("/"));
		if (isFailed(response)) {
			String msg = String.format("OpenSDS Device Info error %d: %s", getErrorCode(response),
					getErrorMessage(response));
			throw new Exception(msg);
		}
		storage = new StorageMO("OpenSDS", response.getString("name"), "", response.getString("status"), "OpenSDS");
	}

	StorageMO getDeviceInfo() {
		return storage;
	}

	JSONObject createVolume(String name, String description, long capacity, String profile) throws Exception {
		JSONObject requestData = new JSONObject();
		requestData.put("name", name);
		requestData.put("size", capacity);
		requestData.put("description", description);
		requestData.put("ProfileId", profile);
		String availabilityZone = "default";
		if (!availabilityZone.isEmpty())
			requestData.put("availabilityZone", availabilityZone);
		JSONObject response = (JSONObject) request.post("/block/volumes", requestData);
		if (isFailed(response)) {
			String msg = String.format("Create volume %s error %d: %s", name, getErrorCode(response),
					getErrorMessage(response));
			throw new Exception(msg);
		}
		return response;
	}

	void deleteVolume(String volumeId) throws Exception {
		JSONObject response = (JSONObject) request.delete(String.format("/block/volumes/%s", volumeId));
		if (response != null) {
			if (isFailed(response)) {
				String msg = String.format("Delete volume %s error %d: %s", volumeId, getErrorCode(response),
						getErrorMessage(response));
				throw new Exception(msg);
			}
		}
	}

	JSONObject attachVolume(String volumeId, ConnectMO connect) throws Exception {
        try {
            JSONObject host = getHost(connect.name);

            if(host.isEmpty()) {
                host = createHost(connect);
            }
            String hostId = host.getString("id");

            JSONObject requestData = new JSONObject();
            requestData.put("volumeId", volumeId);
            requestData.put("hostId", hostId);

            logger.info(String.format("----------OpenSDS Creating Volume Attachment for Volume %s "
                        + "to Host %s----------", volumeId, hostId));

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
        catch (Exception e) {
            logger.error(String.format("Error in creating Volume Attachment, Error Message is: %s", e));
            throw new JSONException("Error in creating Volume Attachment ", e);
        }
    }
	JSONObject getHost(String hostName) throws Exception {
        logger.info(String.format("----------OpenSDS Getting Host %s----------", hostName));

        JSONArray response = (JSONArray)request.get(String.format("/host/hosts?hostName=%s", hostName));
        logger.debug(String.format("OpenSDS Getting Host %s Response: %s", hostName,
                    response));

        if (response.isEmpty()) {
            String msg = String.format("No Host Found");
            logger.info(String.format("OpenSDS Get Host for %s Error: %s", hostName, msg));
            return new JSONObject();
        }

        JSONObject host = (JSONObject) response.get(0);
        return host;
    }
	JSONObject createHost(ConnectMO connect) throws Exception {
        JSONObject requestData = new JSONObject();
        requestData.put("hostName", connect.name);
        requestData.put("ip", connect.initiatorIp);
        requestData.put("osType", connect.osType.toString().toLowerCase());
        requestData.put("accessMode", OPENSDS_HOST_ACCESSMODE);
        String availabilityZone = OPENSDS_AVAILABILITYZONE;

        if(!availabilityZone.isEmpty()) {
			JSONArray availabilityZones = new JSONArray();
			availabilityZones.put(availabilityZone);
			requestData.put("availabilityZones", availabilityZones);
		}

		JSONObject initiatorData = new JSONObject();
		initiatorData.put("portName", connect.iscsiInitiator);
		initiatorData.put("protocol", connect.attachProtocol.toString().toLowerCase());

		JSONArray initiators = new JSONArray();
		initiators.put(initiatorData);
		requestData.put("initiators", initiators);

		logger.info("----------OpenSDS Creating Host----------");

        JSONObject response = (JSONObject)request.post("/host/hosts", requestData);
        logger.debug(String.format("OpenSDS Create Host Response: %s", response));

        if (isFailed(response)) {
            String msg = String.format("Create Host %s Error %d: %s",
                    connect.name, getErrorCode(response), getErrorMessage(response));
            logger.error(msg);
            throw new Exception(msg);
        }

        logger.info("OpenSDS Host Created.");
        return response;
    }


	JSONObject getVolume(String volumeId) throws Exception {

		JSONObject response = (JSONObject) request.get(String.format("/block/volumes/%s", volumeId));
		if (response.isEmpty()) {
			String msg = String.format("List Volume for WWN error: No Volumes Found");
			throw new Exception(msg);
		}
		return response;
	}

	void expandVolume(String volumeId, long capacity) throws Exception {
		JSONObject requestData = new JSONObject();
		requestData.put("newSize", capacity);

		JSONObject response = (JSONObject) request.post(String.format("/block/volumes/%s/resize", volumeId),
				requestData);
		if (isFailed(response)) {
			String msg = String.format("Expand Volume %s error %d: %s", volumeId, getErrorCode(response),
					getErrorMessage(response));
			throw new Exception(msg);
		}
	}

}