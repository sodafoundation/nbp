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

import java.io.IOException;
import java.util.List;

import org.apache.log4j.Level;
import org.apache.log4j.Logger;

import org.opensds.storage.vro.plugin.adapter.opensds.OpenSDSStorageRepository;
import org.opensds.storage.vro.plugin.adapter.opensds.model.StorageMO;

import ch.dunes.util.EncryptHelper;

public class Configuration {
	private static final Logger LOG = Logger.getLogger(Configuration.class);

	private static final String PACKAGE_NAME = "org.opensds.storage";

	// private SysInfoManager sysInfoManager = new SysInfoManager();

	private String id;

	private String name;

	public Configuration(String id, String name) {
		super();
		this.id = id;
		this.name = name;
	}

	public String getId() {
		return id;
	}

	public void setId(String id) {
		this.id = id;
	}

	public String getName() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}

	public String test() {
		return "found";
	}

	/**
	 * register OpenSDS
	 * 
	 * @param arrayName    OpenSDS name
	 * @param hostname     OpenSDS ip address
	 * @param port         OpenSDS port
	 * @param username     OpenSDS username
	 * @param password     OpenSDS password
	 * @param checkCert    check cert or not
	 * @param productModel OpenSDS product model
	 * @throws Exception
	 * @throws StorageCommonException
	 * @throws RestException
	 */
	public void register(String arrayName, String hostname, String port, String username, String password,
			boolean authEnabled, String productModel) throws Exception {
		String uniqId = hostname + "-" + port;
		OpenSDSInfo arrayInfo = new OpenSDSInfo(uniqId, hostname, port);
		arrayInfo.setArrayName(arrayName);
		arrayInfo.setHostName(hostname);
		arrayInfo.setPort(port);
		arrayInfo.setUsername(username);
		arrayInfo.setPassword(password);
		arrayInfo.setauthEnabled(authEnabled);
		arrayInfo.setProductModel(productModel);
		if (LOG.isInfoEnabled()) {
			LOG.info("register:::" + arrayInfo);
		}
		OpenSDS opensds = new OpenSDS();
		opensds.login(arrayInfo);
		StorageMO storage = opensds.getDeviceInfo();
		arrayInfo.setProductName(storage.name);
		arrayInfo.setProductVersion("V1");
		arrayInfo.setProductSN(storage.sn);

		OpenSDSStorageRepository.getUniqueInstance().updateProperties(arrayInfo, OpenSDSStorageRepository.ADD);
	}

	/**
	 * unregister OpenSDS
	 * 
	 * @param opensdsInfo OpenSDS info
	 * @throws IOException
	 */
	public void unregister(OpenSDSInfo opensdsInfo) throws IOException {
		if (LOG.isInfoEnabled()) {
			LOG.info("unregister:::" + opensdsInfo);
		}
		OpenSDSStorageRepository.getUniqueInstance().updateProperties(opensdsInfo, OpenSDSStorageRepository.REMOVE);
	}

	/**
	 * Open DEBUG model
	 * 
	 * @throws StorageCommonException
	 */
	public void debugOn() {
		LOG.info("debug On");
		Logger log = org.apache.log4j.Logger.getLogger(PACKAGE_NAME);
		if (log != null) {
			log.setLevel(Level.DEBUG);
		} else {

		}
	}

	/**
	 * Close DEBUG model
	 * 
	 * @throws StorageCommonException
	 */
	public void debugOff() {
		LOG.info("debug Off");
		Logger log = org.apache.log4j.Logger.getLogger(PACKAGE_NAME);
		if (log != null) {
			log.setLevel(Level.INFO);
		} else {

		}
	}

	/**
	 * Get all OpenSDSs info
	 * 
	 * @return List<OpenSDSInfo>
	 */
	public List<OpenSDSInfo> getOpenSDSInfos() {
		return OpenSDSStorageRepository.getUniqueInstance().getOpenSDSInfos();
	}
}
