package org.opensds.vmware.ngc.service;

import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;

public interface DatastoreService {

    ResultInfo<Object> createDatastore(ManagedObjectReference[] hostMo, ServerInfo serverInfo, DatastoreInfo datastoreInfo);
}
