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

import org.opensds.storage.vro.plugin.adapter.opensds.OpenSDSStorageEventHandler;

public class BaseArrayInfo {
	@Override
	public int hashCode() {
		return this.id.hashCode();
	}

	@Override
	public boolean equals(Object obj) {
		if (obj instanceof BaseArrayInfo) {
			BaseArrayInfo baseArrayInfo = (BaseArrayInfo) obj;
			if (baseArrayInfo.getId().equals(this.id)) {
				return true;
			}
		}
		return false;
	}

	/**
	 * OpenSDSStorage event handler
	 */
	protected OpenSDSStorageEventHandler eventHandler;

	private String id;

	private String arrayName;

	private String hostName;

	private String port;

	private String username;

	private String password;

	private String productModel;

	public String getProductModel() {
		return productModel;
	}

	public void setProductModel(String productModel) {
		this.productModel = productModel;
	}

	public String getArrayName() {
		return arrayName;
	}

	public void setArrayName(String arrayName) {
		this.arrayName = arrayName;
	}

	public String getProductName() {
		return productName;
	}

	public void setProductName(String productName) {
		this.productName = productName;
	}

	public String getProductVersion() {
		return productVersion;
	}

	public void setProductVersion(String productVersion) {
		this.productVersion = productVersion;
	}

	private String productName;

	private String productVersion;

	public BaseArrayInfo() {
		super();
	}

	public BaseArrayInfo(String id, String hostName, String port) {
		super();
		this.id = id;
		this.hostName = hostName;
		this.port = port;
	}

	public String getId() {
		return id;
	}

	/**
	 * get OpenSDS URL
	 * 
	 * @return URL
	 */
	public String getURL() {
		return "https://" + hostName + ":" + port;
	}

	public void setId(String id) {
		this.id = id;
	}

	public String getHostName() {
		return hostName;
	}

	public void setHostName(String hostname) {
		this.hostName = hostname;
	}

	public String getPort() {
		return port;
	}

	public void setPort(String port) {
		this.port = port;
	}

	public String getUsername() {
		return username;
	}

	public void setUsername(String username) {
		this.username = username;
	}

	public String getPassword() {
		return this.password;
	}

	public void setPassword(String password) {
		this.password = password;
	}

	@Override
	public String toString() {
		return "OpenSDSInfo [id=" + id + ", arrayName=" + arrayName + ", hostName=" + hostName + ", port=" + port
				+ ", productName=" + productName + ", productVersion=" + productVersion + ", productModel="
				+ productModel + "]";
	}
}
