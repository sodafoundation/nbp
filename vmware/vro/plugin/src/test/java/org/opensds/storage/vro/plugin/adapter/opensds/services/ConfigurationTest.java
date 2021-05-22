package org.opensds.storage.vro.plugin.adapter.opensds.services;

import org.junit.jupiter.api.Test;

class ConfigurationTest {

	@Test
	void testRegister() throws Exception {
		Configuration conf = new Configuration("testID", "opensds");
		conf.register("OpenSDS_Test", "127.0.0.1", "50040", "admin", "opensds@123", true, "V1");
	}

}
