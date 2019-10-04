package org.opensds.vmware.ngc.adapters.opensds;

import java.util.List;

import org.opensds.vmware.ngc.models.SnapshotMO;
import org.opensds.vmware.ngc.models.VolumeMO;

public class Main {
	public static void main(String args[]) throws Exception {
		OpenSDS openSDS = new OpenSDS("opensds");
		openSDS.login("192.168.20.159", 50040, "admin", "opensds@123");
		List<VolumeMO> list = openSDS.listVolumes();
		//VolumeMO vol =openSDS.queryVolumeByID("688d5620cb544e49b8d759c0c517c16c1");
		//System.out.println(vol.name);
		for(VolumeMO volume : list) {
			System.out.println(volume.id);
		}
	}

}
