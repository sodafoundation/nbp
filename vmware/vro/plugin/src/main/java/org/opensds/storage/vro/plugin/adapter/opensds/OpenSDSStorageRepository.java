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

package org.opensds.storage.vro.plugin.adapter.opensds;

import java.io.File;
import java.io.IOException;
import java.util.List;
import java.util.Properties;
import java.util.Vector;

import org.apache.log4j.Logger;

import org.opensds.storage.vro.plugin.adapter.opensds.services.OpenSDSInfo;

import ch.dunes.vso.sdk.helper.SDKHelper;

public class OpenSDSStorageRepository {

	public static final String REMOVE = "REMOVE";

	public static final String ADD = "ADD";

	private static final Logger LOG = Logger.getLogger(OpenSDSStorageRepository.class);

	private static final Object LOCK = new Object();

	private static final String MODULE_NAME = "OpenSDS-storage";

	private static final String CONNECTOR_STR = ".";

	private static final String HOSTS_LIST_PROP_KEY = "HOST_LIST";

	private static final String HOST_SEPARATER = ";";

	private static final String HOSTNAME_PROP_KEY = "HOSTNAME";

	private static final String PORT_PROP_KEY = "PORT";

	private static final String USERNAME_PROP_KEY = "USERNAME";

	private static final String PAD_PROP_KEY = "PASSWORD";

	private static final String CHECK_CERT = "CHECK_CERT";

	private static final String STORAGE_ARRAY_NAME_PROP_KEY = "STORAGE_ARRAY_NAME";

	private static final String PRODUCT_NAME_PROP_KEY = "PRODUCT_NAME";

	private static final String PRODUCT_VERSION_PROP_KEY = "PRODUCT_VERSION";

	private static final String PRODUCT_MODEL_PROP_KEY = "PRODUCT_MODEL";

	private static final String PRODUCT_SN_PROP_KEY = "PRODUCT_SN";

	private static OpenSDSStorageRepository uniqueInstance;

	private final Object LOCK_CLEAN_CACHE = new Object();

	private final Object LOCK_UPDATE_PROPERTIES = new Object();

	private OpenSDSStorageEventHandler eventHandler;

	private List<OpenSDSInfo> arrays = new Vector<OpenSDSInfo>();

	private OpenSDSStorageRepository() {
		arrays = getAllOpenSDSInfo();
		this.eventHandler = new OpenSDSStorageEventHandler();
	}

	/**
	 * add OpenSDS Controller
	 * 
	 * @param arrayInfo OpenSDS controller info
	 * @throws IOException
	 */
	public void addOpenSDS(OpenSDSInfo arrayInfo) throws IOException {
		if (LOG.isInfoEnabled()) {
			LOG.info("enter method addOpenSDS" + arrayInfo.getId());
		}
		String path = SDKHelper.getConfigurationPathForPluginName(MODULE_NAME);
		Properties prop = null;
		String uniqId = arrayInfo.getHostName() + "-" + arrayInfo.getPort();

		if (!new File(path).exists()) {
			prop = new Properties();
			prop.put(HOSTS_LIST_PROP_KEY, uniqId);
		} else {
			prop = SDKHelper.loadPropertiesForPluginName(MODULE_NAME);
			String oldHostListStr = prop.getProperty(HOSTS_LIST_PROP_KEY);
			if (oldHostListStr != null && !oldHostListStr.trim().equals("")) {
				if (oldHostListStr.contains(uniqId)) {
					LOG.warn(uniqId + " already registerred!");
				} else {
					prop.put(HOSTS_LIST_PROP_KEY, oldHostListStr + HOST_SEPARATER + uniqId);
				}
			} else {
				// New Registration
				prop.put(HOSTS_LIST_PROP_KEY, uniqId);
			}
		}
		prop.put(uniqId + CONNECTOR_STR + STORAGE_ARRAY_NAME_PROP_KEY, arrayInfo.getArrayName());
		prop.put(uniqId + CONNECTOR_STR + HOSTNAME_PROP_KEY, arrayInfo.getHostName());
		prop.put(uniqId + CONNECTOR_STR + PORT_PROP_KEY, arrayInfo.getPort());
		prop.put(uniqId + CONNECTOR_STR + USERNAME_PROP_KEY, arrayInfo.getUsername());
		prop.put(uniqId + CONNECTOR_STR + PAD_PROP_KEY, arrayInfo.getPassword());
		prop.put(uniqId + CONNECTOR_STR + CHECK_CERT, String.valueOf(arrayInfo.getauthEnabled()));
		prop.put(uniqId + CONNECTOR_STR + PRODUCT_NAME_PROP_KEY, String.valueOf(arrayInfo.getProductName()));
		prop.put(uniqId + CONNECTOR_STR + PRODUCT_VERSION_PROP_KEY, String.valueOf(arrayInfo.getProductVersion()));
		prop.put(uniqId + CONNECTOR_STR + PRODUCT_MODEL_PROP_KEY, String.valueOf(arrayInfo.getProductModel()));
		prop.put(uniqId + CONNECTOR_STR + PRODUCT_SN_PROP_KEY, String.valueOf(arrayInfo.getProductSN()));
		SDKHelper.savePropertiesForPluginName(prop, MODULE_NAME);
		// to avoid duplicate info in memory cache
		arrays.remove(arrayInfo);
		arrays.add(arrayInfo);
		LOG.info("exit method addOpenSDS");
	}

	private void removeOpenSDS(OpenSDSInfo arrayInfo) throws IOException {
		if (LOG.isInfoEnabled()) {
			LOG.info("enter method removeOpenSDS" + arrayInfo.getId());
		}
		String path = SDKHelper.getConfigurationPathForPluginName(MODULE_NAME);
		String uniqId = arrayInfo.getHostName() + "-" + arrayInfo.getPort();
		if (!new File(path).exists()) {
			LOG.error("registration file not exist!");
			arrays.remove(arrayInfo);
			return;
		} else {
			Properties prop = SDKHelper.loadPropertiesForPluginName(MODULE_NAME);
			String oldHostListStr = prop.getProperty(HOSTS_LIST_PROP_KEY);
			if (oldHostListStr == null || oldHostListStr.trim().equals("")) {
				LOG.warn("registration hostList not exist!");
				arrays.remove(arrayInfo);
				return;
			}
			if (!oldHostListStr.contains(uniqId)) {
				LOG.warn("can not find registration info for " + uniqId);
				arrays.remove(arrayInfo);
				return;
			}
			String[] oldHostList = oldHostListStr.split(HOST_SEPARATER);
			if (oldHostList.length == 0) {
				LOG.warn("registration HostListStr count is 0");
				arrays.remove(arrayInfo);
				return;
			}
			StringBuffer newHostListBuff = new StringBuffer();
			int counter = 0;
			for (String key : oldHostList) {
				if (key.equals(uniqId)) {
					prop.remove(uniqId + CONNECTOR_STR + STORAGE_ARRAY_NAME_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + HOSTNAME_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + PORT_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + USERNAME_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + PAD_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + CHECK_CERT);
					prop.remove(uniqId + CONNECTOR_STR + PRODUCT_NAME_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + PRODUCT_VERSION_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + PRODUCT_MODEL_PROP_KEY);
					prop.remove(uniqId + CONNECTOR_STR + PRODUCT_SN_PROP_KEY);
					continue;
				}

				if (!key.equals("")) {
					if (counter > 0) {
						newHostListBuff.append(HOST_SEPARATER);
					}
					newHostListBuff.append(key);
				}
				counter++;
			}
			prop.put(HOSTS_LIST_PROP_KEY, newHostListBuff.toString());
			SDKHelper.savePropertiesForPluginName(prop, MODULE_NAME);
			for (OpenSDSInfo array : arrays) {
				if (array.equals(arrayInfo)) {
					array.clearAll();
				}
			}
		}
		arrays.remove(arrayInfo);
		if (LOG.isInfoEnabled()) {
			LOG.info("exit method removeOpenSDS");
		}
	}

	private List<OpenSDSInfo> getAllOpenSDSInfo() {
		List<OpenSDSInfo> list = new Vector<OpenSDSInfo>();
		String path = SDKHelper.getConfigurationPathForPluginName(MODULE_NAME);
		if (new File(path).exists()) {
			try {
				Properties prop = SDKHelper.loadPropertiesForPluginName(MODULE_NAME);
				String oldHostListStr = prop.getProperty(HOSTS_LIST_PROP_KEY);
				if (oldHostListStr != null && !oldHostListStr.equals("")) {
					String[] oldHostList = oldHostListStr.split(HOST_SEPARATER);
					for (String key : oldHostList) {
						list.add(createOpenSDSInfo(prop, key));
					}
				}
			} catch (IOException e) {
				LOG.error("getAllOpenSDSInfo-IOException" + e.getMessage());
			}
		}
		return list;
	}

	/**
	 * get OpenSDS by id
	 * 
	 * @param id OpenSDS id
	 * @return OpenSDS OpenSDS info
	 */
	public OpenSDSInfo getOpenSDSInfoById(String id) {
		if (arrays.size() != 0) {
			String realId = id;
			if (id.contains(",")) {
				realId = id.split(",")[0];
			}
			for (OpenSDSInfo openSDSInfo : arrays) {
				if (openSDSInfo.getId().equals(realId)) {
					return openSDSInfo;
				}
			}
		}
		return new OpenSDSInfo();
	}

	private OpenSDSInfo createOpenSDSInfo(Properties prop, String uniqId) {
		OpenSDSInfo array = new OpenSDSInfo();
		array.setArrayName(prop.getProperty(uniqId + CONNECTOR_STR + STORAGE_ARRAY_NAME_PROP_KEY));
		array.setHostName(prop.getProperty(uniqId + CONNECTOR_STR + HOSTNAME_PROP_KEY));
		array.setId(uniqId);
		array.setPort(prop.getProperty(uniqId + CONNECTOR_STR + PORT_PROP_KEY));
		array.setUsername(prop.getProperty(uniqId + CONNECTOR_STR + USERNAME_PROP_KEY));
		array.setPassword(prop.getProperty(uniqId + CONNECTOR_STR + PAD_PROP_KEY));
		array.setauthEnabled(Boolean.valueOf(prop.getProperty(uniqId + CONNECTOR_STR + CHECK_CERT)));
		array.setProductName(prop.getProperty(uniqId + CONNECTOR_STR + PRODUCT_NAME_PROP_KEY));
		array.setProductVersion(prop.getProperty(uniqId + CONNECTOR_STR + PRODUCT_VERSION_PROP_KEY));
		array.setProductModel(prop.getProperty(uniqId + CONNECTOR_STR + PRODUCT_MODEL_PROP_KEY));
		array.setProductSN(prop.getProperty(uniqId + CONNECTOR_STR + PRODUCT_SN_PROP_KEY));
		return array;
	}

	/**
	 * update OpenSDS info
	 * 
	 * @param opesdsInfo OpenSDS info
	 * @param actionType add or remove OpenSDS info
	 * @throws IOException
	 */
	public void updateProperties(OpenSDSInfo opesdsInfo, String actionType) throws IOException {
		synchronized (LOCK_UPDATE_PROPERTIES) {
			if (actionType.equals(ADD)) {
				addOpenSDS(opesdsInfo);
			} else {
				removeOpenSDS(opesdsInfo);
			}
		}
	}

	public static OpenSDSStorageRepository getUniqueInstance() {
		synchronized (LOCK) {
			if (uniqueInstance == null) {
				uniqueInstance = new OpenSDSStorageRepository();
			}
			return uniqueInstance;
		}
	}

	/**
	 * get OpenSDS info by range
	 * 
	 * @return List<OpenSDSInfo>
	 */
	public List<OpenSDSInfo> getOpenSDSInfos() {
		return arrays;
	}

	/**
	 * register listener
	 * 
	 * @param listener IOpenSDSStorageListener
	 */
	public void registerListener(IOpenSDSStorageListener listener) {
		this.eventHandler.registerListener(listener);
	}

	/**
	 * unregister listener
	 * 
	 * @param listener IOpenSDSStorageListener
	 */
	public void unregisterListener(IOpenSDSStorageListener listener) {
		this.eventHandler.unregisterListener(listener);
	}
}
