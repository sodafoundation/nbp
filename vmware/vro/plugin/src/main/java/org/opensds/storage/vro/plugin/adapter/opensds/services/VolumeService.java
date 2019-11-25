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

import org.apache.log4j.Logger;
import org.opensds.storage.vro.plugin.adapter.opensds.model.VolumeMO;

public class VolumeService {
	private static final Logger log = Logger.getLogger(VolumeService.class);

	public void createVolume(OpenSDSInfo openSDSInfo, String name, String description, long capacity, String profile)
			throws Exception {

		OpenSDS opensds = new OpenSDS();
		opensds.login(openSDSInfo);
		log.info("Logged in to OpenSDS");
		VolumeMO volume = opensds.createVolume(name, description, capacity, profile);
		log.info("Volume " + volume.id + "is created");

	}

	public void deleteVolume(OpenSDSInfo openSDSInfo, String id) throws Exception {

		OpenSDS opensds = new OpenSDS();
		opensds.login(openSDSInfo);
		log.info("Logged in to OpenSDS");
		opensds.deleteVolume(id);
		log.info("Volume " + id + "is deleted");

	}

}
