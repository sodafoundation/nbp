package org.opensds.storage.vro.plugin.adapter.opensds.services;

import org.junit.jupiter.api.Test;

class OpenSDSInfoTest {

	@Test
	void testOpenSDSInfo() {
		OpenSDSInfo OpenSDSInfo = new OpenSDSInfo("testId", "testHostName", "testPort");
		OpenSDSInfo.hashCode();
		OpenSDSInfo.getHostName();
		OpenSDSInfo.getURL();
		OpenSDSInfo.getProductSN();
		OpenSDSInfo.toString();
	}

	@Test
	void testEqualsObject() {
		OpenSDSInfo OpenSDSInfo = new OpenSDSInfo("testId", "testHostName", "testPort");
		OpenSDSInfo obj = new OpenSDSInfo("testId", "testHostName", "testPort");
		OpenSDSInfo.equals(obj);
	}

	@Test
	void testNotEqualsObject() {
		OpenSDSInfo OpenSDSInfo = new OpenSDSInfo();
		OpenSDSInfo obj = new OpenSDSInfo("testId", "testHostName", "testPort");
		OpenSDSInfo.equals(obj);
	}

	@Test
	void testBaseArrayInfo() {
		BaseArrayInfo baseArray = new BaseArrayInfo("testID", "testHost", "testPort");
		baseArray.getId();
		baseArray.getHostName();
		baseArray.getPassword();
		baseArray.getURL();
		baseArray.getArrayName();
		baseArray.toString();
		baseArray.getProductModel();
		baseArray.getProductName();
		baseArray.getProductVersion();

	}

}
