package org.opensds.vmware.ngc.service.impl;

import static org.junit.Assert.fail;

import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.Properties;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.opensds.vmware.ngc.adapter.DeviceDataAdapter;
import org.opensds.vmware.ngc.adapters.opensds.OpenSDS;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.dao.VolumesRepository;
import org.opensds.vmware.ngc.dao.impl.DeviceRepositoryImpl;
import org.opensds.vmware.ngc.dao.impl.VolumesRepositoryImpl;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.service.DeviceService;
import org.opensds.vmware.ngc.service.impl.DeviceServiceImpl;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.test.context.ContextConfiguration;
import org.springframework.test.context.junit4.SpringJUnit4ClassRunner;

import com.vmware.vise.vim.data.VimObjectReferenceService;

@RunWith(SpringJUnit4ClassRunner.class)
@ContextConfiguration
public class DeviceServiceTest {

	private String openSdsIP;
	private String userName;
	private String password;
	private int port;
	private String esxIP;
	private String esxIQN;
	private static String deviceID = "123456";

	@Configuration
	static class DeviceServiceConfiguration {

		@Bean
		public DeviceService deviceService() {
			DeviceService temp = new DeviceServiceImpl();
			return temp;
		}

		@Bean
		public DeviceRepository deviceRepository() {
			return new DeviceRepositoryImpl();
		}

		@Bean
		public VimObjectReferenceService vimObjectReferenceService() {

			return new VimObjectReferenceService() {

				@Override
				public String getUid(Object arg0, boolean arg1) {
					// TODO Auto-generated method stub
					return deviceID;
				}

				@Override
				public String getUid(Object arg0) {
					// TODO Auto-generated method stub
					return deviceID;
				}

				@Override
				public String getServerGuid(Object arg0) {
					// TODO Auto-generated method stub
					return null;
				}

				@Override
				public String getResourceObjectType(Object arg0) {
					// TODO Auto-generated method stub
					return null;
				}

				@Override
				public Object getReference(String arg0, boolean arg1) {
					// TODO Auto-generated method stub
					return null;
				}

				@Override
				public Object getReference(String arg0) {
					// TODO Auto-generated method stub
					return null;
				}

				@Override
				public String getValue(Object arg0) {
					// TODO Auto-generated method stub
					return null;
				}

				@Override
				public Object getReference(String arg0, String arg1, String arg2) {
					// TODO Auto-generated method stub
					return new OpenSDS("OpenSDS");
				}
			};
		}

		@Bean
		public VolumesRepository volumesRepository() {

			return new VolumesRepositoryImpl();
		}
	}

	@Autowired
	private DeviceService deviceService;

	@Autowired
	private DeviceRepository deviceRepository;

	@Autowired
	private VimObjectReferenceService vor;

	@Autowired
	private VolumesRepository volumesRepository;

	@Before
	public void setup() {
		ResultInfo<Object> resultInfo = new ResultInfo<Object>();
	}

	@Test()
	public void testDeviceServiceIT() throws Exception {
		readDefaultConfig();
		DeviceInfo deviceInfo = new DeviceInfo(openSdsIP, userName, password, port, "OpenSDS", "none");
		ResultInfo<Object> resultInfo = deviceService.add(deviceInfo);
		if (resultInfo == null) {
			fail("Device addition failure");
		}
		DeviceServiceConfiguration dsc = new DeviceServiceConfiguration();
		Object deviceReference = dsc.vimObjectReferenceService().getReference(DeviceDataAdapter.DEVICE_TYPE,
				deviceInfo.ip, null);
		DeviceInfo deviceInfo_updated = new DeviceInfo(openSdsIP, userName, password, port, "OpenSDS", "updated");
		ResultInfo<Object> resultInfo_updated = deviceService.update(deviceReference, deviceInfo_updated);
		if (resultInfo_updated == null) {
			fail("Device update  failure");
		}
		resultInfo = deviceService.get(deviceReference);
		if (resultInfo == null) {
			fail("Device get failure");
		}
		resultInfo = deviceService.getAllDeviceType();
		if (resultInfo == null) {
			fail("Device getAllDeviceType failure");
		}
		resultInfo = deviceService.getDeviceBlockPools(deviceID);
		if (resultInfo == null) {
			fail("Device getDeviceBlockPools failure");
		}
		resultInfo = deviceService.getList();
		if (resultInfo == null) {
			fail("Device getList failure");
		}
		// unauthorized rest
		DeviceInfo deviceInfo_auth_fail = new DeviceInfo(openSdsIP, "wrong_userid", password, port, "OpenSDS", "none");
		deviceService.update(deviceReference, deviceInfo_auth_fail);
		resultInfo = deviceService.getDeviceBlockPools(deviceID);
		deviceService.delete(deviceReference);
		// add new device again
		resultInfo = deviceService.add(deviceInfo);
		if (resultInfo == null) {
			fail("Device addition failure");
		}

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
