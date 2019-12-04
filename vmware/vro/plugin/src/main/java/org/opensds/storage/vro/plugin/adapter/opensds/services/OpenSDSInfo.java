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

import java.util.List;
import java.util.Objects;
import java.util.Vector;

public class OpenSDSInfo extends BaseArrayInfo {
	@Override
	public boolean equals(Object obj) {
		if (obj instanceof OpenSDSInfo) {
			OpenSDSInfo openSDSInfo = (OpenSDSInfo) obj;
			if (openSDSInfo.getId().equals(getId())) {
				return true;
			}
		}
		return false;
	}

	@Override
	public int hashCode() {
		return Objects.hash(super.hashCode(), productSN, authEnabled);
	}

	private String productSN;
	private boolean authEnabled = false;

	public String getProductSN() {
		return productSN;
	}

	public void setProductSN(String productSN) {
		this.productSN = productSN;
	}

	public boolean getauthEnabled() {
		return authEnabled;
	}

	public void setauthEnabled(boolean authEnabled) {
		this.authEnabled = authEnabled;
	}

	public void clearAll() {

	}

	@Override
	public String getURL() {
		return "https://" + getHostName() + ":" + getPort() + "/deviceManager/rest";
	}

	@Override
	public String toString() {
		return "OpenSDSInfo [id=" + getId() + ", arrayName=" + getArrayName() + ", hostName=" + getHostName()
				+ ", port=" + getPort() + ", authEnabled=" + getauthEnabled() + ", productName=" + getProductName()
				+ ", productVersion=" + getProductVersion() + ", productModel=" + getProductModel() + "]";
	}

	public OpenSDSInfo() {
		super();
	}

	public OpenSDSInfo(String id, String hostName, String port) {
		super();
		setId(id);
		setHostName(hostName);
		setPort(port);
	}
}
