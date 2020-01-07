package org.opensds.storage.vro.plugin.adapter.opensds.services;

import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.Properties;

import org.junit.jupiter.api.Test;

class VolumeServiceTest {
	private OpenSDSInfo arrayInfo = new OpenSDSInfo("OpenSDS", "test", "50040");
	private String profileId;
	private String esxIP;
	private String esxIQN;
	private String volId;

	public void readDefaultConfig() throws IOException {
		System.out.println("Working Directory = " + System.getProperty("user.dir"));
		Properties p = new Properties();
		p.load(new FileReader(new File(".\\src\\test\\resources\\config.properties")));
		setOpenSDSInfo(p);
		profileId = p.getProperty("ProfileID");
		esxIP = p.getProperty("ESX_IP");
		esxIQN = p.getProperty("ESX_IQN");
		volId = p.getProperty("VOLUME_ID");
	}

	void setOpenSDSInfo(Properties p) {
		arrayInfo.setArrayName("OpenSDS");
		arrayInfo.setHostName(p.getProperty("HostIP"));
		arrayInfo.setPort(p.getProperty("Port"));
		arrayInfo.setPassword(p.getProperty("Password"));
		arrayInfo.setUsername(p.getProperty("UserName"));
		arrayInfo.setauthEnabled(p.getProperty("AuthEnabled").equals("true") ? true : false);
		arrayInfo.setProductModel("V1");
	}

	@Test
	void testCreateVolume() throws Exception {
		readDefaultConfig();
		VolumeService volumeService = new VolumeService();
		volumeService.createVolume(arrayInfo, "test_script_vol", "test_script_vol", 1, profileId);
	}

	@Test
	void testcreateAndAttachVolume() throws Exception {
		VolumeService volumeService = new VolumeService();
		readDefaultConfig();
		volumeService.createAndAttachVolume(arrayInfo, "attach_vol", "attach volume test", 1, profileId, esxIP, esxIQN);
	}

	@Test
	void testDeleteVolume() throws Exception {
		VolumeService volumeService = new VolumeService();
		readDefaultConfig();
		volumeService.deleteVolume(arrayInfo, volId);
	}

	@Test
	void testexpandVolume() throws Exception {
		VolumeService volumeService = new VolumeService();
		readDefaultConfig();
		volumeService.expandVolume(arrayInfo, volId, 2);
	}

}
