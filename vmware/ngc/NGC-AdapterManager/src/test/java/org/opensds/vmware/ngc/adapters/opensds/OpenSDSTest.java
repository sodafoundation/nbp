package org.opensds.vmware.ngc.adapters.opensds;

import static org.junit.jupiter.api.Assertions.*;

import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.List;
import java.util.Properties;

import org.junit.jupiter.api.Test;
import org.opensds.vmware.ngc.models.ALLOC_TYPE;
import org.opensds.vmware.ngc.models.ATTACH_MODE;
import org.opensds.vmware.ngc.models.ATTACH_PROTOCOL;
import org.opensds.vmware.ngc.models.ConnectMO;
import org.opensds.vmware.ngc.models.HOST_OS_TYPE;
import org.opensds.vmware.ngc.models.POOL_TYPE;
import org.opensds.vmware.ngc.models.SnapshotMO;
import org.opensds.vmware.ngc.models.StoragePoolMO;
import org.opensds.vmware.ngc.models.VolumeMO;

class OpenSDSTest {
	private String openSdsIP;
	private String userName;
	private String password;
	private int port;
	private String esxIP;
	private String esxIQN;
	private static int VOLUME_SIZE=1;
	private static int VOLUME_EXPAND_SIZE=2;
	private static int SLEEP_DURATION=2000;

	@Test
	void ITtestOpenSDS() throws Exception {
		OpenSDS osds = new OpenSDS("OpenSDS");
		readDefaultConfig();
		osds.login(openSdsIP, port, userName, password);
		osds.getDeviceInfo();
		List<StoragePoolMO> pools = osds.listStoragePools();
		StoragePoolMO pool = SelectBlockPool(pools);
		if (pool == null) {
			fail("List Pool  Failure");
		}
		VolumeMO volume = osds.createVolume("test_volume", "test volume creation", ALLOC_TYPE.THIN, VOLUME_SIZE*UNIT_TYPE.GB.getUnit(),
				pool.id);
		Thread.sleep(SLEEP_DURATION);
		if (volume == null) {
			fail("Volume Creation Failure");
		}
		osds.expandVolume(volume.id, (long) VOLUME_EXPAND_SIZE * UNIT_TYPE.GB.getUnit());
		Thread.sleep(SLEEP_DURATION);
		osds.createVolumeSnapshot(volume.id, "test_volume_snap");
		Thread.sleep(SLEEP_DURATION);
		List<VolumeMO> volumes = osds.listVolumes();
		if (volumes == null) {
			fail("List Volumes  Failure");
		}
		List<VolumeMO> pool_volumes = osds.listVolumes(pool.id);
		if (pool_volumes == null) {
			fail("List Volumes with PoolID  Failure");
		}
		List<SnapshotMO> snapshots = osds.listSnapshot(volume.id);
		if (snapshots == null) {
			fail("List Sanpshot  Failure");
		}
		StoragePoolMO pool1 = osds.getStoragePool(pool.id);
		if (pool1 == null) {
			fail("Get Pool  Failure");
		}
		List<VolumeMO> filterVolume = osds.listVolumes("DurableName", volume.wwn);
		if (filterVolume == null) {
			fail("List Volumes by DurableName Failed");
		}
		VolumeMO filterVolume1 = osds.queryVolumeByID(volume.wwn);
		if (filterVolume1 == null) {
			fail("List Volumes by ID Failed");
		}
		ConnectMO connectMO = new ConnectMO("esx_host", HOST_OS_TYPE.ESXI, esxIQN, esxIP, null, ATTACH_MODE.RW,
				ATTACH_PROTOCOL.ISCSI);
		osds.attachVolume(volume.id, connectMO);
		Thread.sleep(SLEEP_DURATION);
		osds.detachVolume(volume.id, connectMO);
		Thread.sleep(SLEEP_DURATION);
		osds.deleteVolumeSnapshot(snapshots.get(0).id);
		Thread.sleep(SLEEP_DURATION);
		osds.deleteVolume(volume.id);
		osds.logout();
	}

	private StoragePoolMO SelectBlockPool(List<StoragePoolMO> pools) {

		for (StoragePoolMO temp : pools) {
			if (temp.type == POOL_TYPE.BLOCK) {
				return temp;
			}
		}
		return null;
	}

	public void readDefaultConfig() throws IOException {
		Properties p = new Properties();
		p.load(new FileReader(new File(".\\src\\test\\resources\\config.properties")));
		esxIP = p.getProperty("ESX_IP");
		esxIQN = p.getProperty("ESX_IQN");
		openSdsIP = p.getProperty("HostIP");
		userName = p.getProperty("UserName");
		password = p.getProperty("Password");
		port = Integer.parseInt(p.getProperty("Port"));
	}

}
