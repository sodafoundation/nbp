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

import org.json.JSONObject;
import org.opensds.storage.vro.plugin.adapter.opensds.model.*;

class VolumeMOBuilder {
	static public VolumeMO build(JSONObject jsonObject) {
		String wwn = "";
		String name = jsonObject.getString("name");
		String id = jsonObject.getString("id");
		if (jsonObject.has("identifier")) {
			JSONObject identifier = jsonObject.getJSONObject("identifier");
			wwn = identifier.getString("durableName");
		}
		ALLOC_TYPE allocType = ALLOC_TYPE.THIN;
		long capacity = jsonObject.getLong("size");
		return new VolumeMO(name, id, wwn, allocType, capacity);
	}
}

public class OpenSDS {

	RestClient client;

	public OpenSDS() {
		this.client = new RestClient();
	}

	public void login(OpenSDSInfo openSDSInfo) throws Exception {
		client.login(openSDSInfo);
	}

	public void logout() {
		client.logout();
	}

	public StorageMO getDeviceInfo() throws Exception {
		return client.getDeviceInfo();
	}

	public VolumeMO createVolume(String name, String description, long capacity, String profile) throws Exception {

		JSONObject volume = client.createVolume(name, description, capacity, profile);
		return VolumeMOBuilder.build(volume);
	}

	public void deleteVolume(String volumeId) throws Exception {
		client.deleteVolume(volumeId);
	}

	public void attachVolume(String volumeId, ConnectMO connect) throws Exception {
		client.attachVolume(volumeId, connect);
	}

	public VolumeMO queryVolumeByID(String volumeId) throws Exception {
		JSONObject volume = client.getVolume(volumeId);
		return VolumeMOBuilder.build(volume);
	}

	public String getVolumeWWN(String volid) throws Exception {
		VolumeMO volume = queryVolumeByID(volid);

		return volume.wwn;
	}

	public void expandVolume(String volumeId, long capacity) throws Exception {
		client.expandVolume(volumeId, capacity);
	}

}
