/**
 * 
 */

package org.opensds.vmware.ngc.service.impl;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.runners.MockitoJUnitRunner;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.model.datastore.NFSDatastore;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import org.opensds.vmware.ngc.service.impl.DatastoreServiceImpl;
import org.springframework.beans.factory.annotation.Autowired;
import static org.mockito.Mockito.*;

import static org.junit.Assert.assertEquals;

import org.junit.Before;

@RunWith(MockitoJUnitRunner.class)
public class DatastoreTest {
	
	@InjectMocks
	public DatastoreServiceImpl datastoreService;
	
	@Mock
	VMFSDatastore vmfsDataStore;
	
	@Mock
	NFSDatastore nfsDataStore;
	
	 @Mock
	 public DeviceRepository deviceRepository;
	
	
	@Test
	public void create_datastore_test_when_all_param_null(){		
		ResultInfo<Object> resultInfo = datastoreService.create(null, null, null);
		assertEquals(resultInfo.getStatus(), "error");
	}
	
	@Test
	public void create_datastore_test_default_vmfsdatastore(){		
	
		ResultInfo<Object> resultInfo = datastoreService.create(null, null, vmfsDataStore);
		assertEquals(resultInfo.getStatus(), "error");
	}
	
	@Test
	public void create_datastore_test_default_nfsdatastore(){		
	
		ResultInfo<Object> resultInfo = datastoreService.create(null, null, nfsDataStore);
		assertEquals(resultInfo.getStatus(), "error");
	}
	
	
	@Test
	public void create_datastore_test_nfsdatastore_with_empty_storageid(){		
		NFSDatastore nfsDataStoreStoId = new NFSDatastore();
		nfsDataStoreStoId.setStorageId("");
		ResultInfo<Object> resultInfo = datastoreService.create(null, null, nfsDataStoreStoId);
		assertEquals(resultInfo.getStatus(), "error");
	}
	
	@Test
	public void create_datastore_test_nfsdatastore_with_storageid(){	
		DeviceInfo deviceInfo=null;
		NFSDatastore nfsDataStoreStoId = new NFSDatastore();
		nfsDataStoreStoId.setStorageId("storage_1");
		when(deviceRepository.get(nfsDataStoreStoId.getStorageId())).thenReturn(deviceInfo);
		ResultInfo<Object> resultInfo = datastoreService.create(null, null, nfsDataStoreStoId);
		assertEquals(resultInfo.getStatus(), "error");
	}
	
	
}