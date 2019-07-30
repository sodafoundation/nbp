package org.opensds.vmware.ngc.service;

import org.opensds.vmware.ngc.entity.ResultInfo;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;


public interface DataCenterService {

    ResultInfo<Object> getHostListByClusterId(ManagedObjectReference clusterMOR, final ServerInfo serverInfo);

    ResultInfo<Object> getHostListByDataCenterId(ManagedObjectReference datacenterMOR, final ServerInfo serverInfo);
}
