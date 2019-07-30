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

package org.opensds.vmware.ngc.service.impl;

import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.expections.ExpectionHandle;
import org.opensds.vmware.ngc.expections.InactiveSessionException;
import org.opensds.vmware.ngc.model.HostInfo;
import org.opensds.vmware.ngc.service.DataCenterService;
import com.vmware.vim25.*;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;


@Service
public class DataCenterServiceImpl extends VimCommonServiceImpl implements DataCenterService {

    private static final Log logger = LogFactory.getLog(DataCenterServiceImpl.class);

    @Override
    public ResultInfo<Object> getHostListByDataCenterId(
            ManagedObjectReference datacenterMOR, final ServerInfo serverInfo) {

        logger.info("Begin get the HostList info in datacenter");
        ResultInfo<Object> resultInfo = new ResultInfo();
        List<HostInfo> hosts = new ArrayList<HostInfo>();
        try {
            Map<String, Object> propertiesMapGroupList = getMoProperties(datacenterMOR, serverInfo, "hostFolder");
            ManagedObjectReference hostFolder = (ManagedObjectReference) propertiesMapGroupList.get("hostFolder");
            Map<String, Object> propertiesClusterList = getMoProperties(hostFolder, serverInfo, "childEntity");
            ArrayOfManagedObjectReference objectReference = (ArrayOfManagedObjectReference) propertiesClusterList.get("childEntity");
            List<ManagedObjectReference> morList = objectReference.getManagedObjectReference();

            for (ManagedObjectReference clusterMOR : morList) {
                Map<String, Object> propertiesHostList = getMoProperties(clusterMOR, serverInfo, "host");
                ArrayOfManagedObjectReference objectReferenceHosts = (ArrayOfManagedObjectReference) propertiesHostList.get("host");
                List<ManagedObjectReference> morHostSystemList = objectReferenceHosts.getManagedObjectReference();
                hosts.addAll(getHostInCenter(morHostSystemList, serverInfo));
            }

        }catch (Exception e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
            resultInfo.setMsg(e.getMessage());
        }

        resultInfo.setData(hosts);
        return resultInfo;
    }

    @Override
    public ResultInfo<Object> getHostListByClusterId(
            ManagedObjectReference clusterMOR, final ServerInfo serverInfo){

        logger.info("Begin get the HostList info in cluster.");
        ResultInfo<Object> resultInfo = new ResultInfo();
        try {
            Map<String, Object> propertiesHostList = getMoProperties(clusterMOR, serverInfo, "host");
            ArrayOfManagedObjectReference objectReferenceHosts = (ArrayOfManagedObjectReference) propertiesHostList.get("host");
            List<ManagedObjectReference> morHostSystemList = objectReferenceHosts.getManagedObjectReference();
            List<HostInfo> hosts  = getHostInCenter(morHostSystemList, serverInfo);
            resultInfo.setData(hosts);
        }catch (Exception e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
            resultInfo.setMsg(e.getMessage());
        }
        resultInfo.setData(new ArrayList<HostInfo>());
        return resultInfo;
    }


    private List<HostInfo> getHostInCenter(
            final List<ManagedObjectReference> morHostSystemList,
            final ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        List<HostInfo> hosts = new ArrayList<HostInfo>();
        for(ManagedObjectReference hostMOR: morHostSystemList){
            HostInfo host = new HostInfo();
            String hostname = hostMOR.getValue();
            Map<String, Object> propertiesMapHosts = getMoProperties(hostMOR, serverInfo, "name", "runtime");
            String hostIp = (String) propertiesMapHosts.get("name");
            HostRuntimeInfo hostRuntimeInfo =  (HostRuntimeInfo) propertiesMapHosts.get("runtime");
            host.setName(hostname);
            host.setIp(hostIp);
            host.setConnectedType(hostRuntimeInfo.getConnectionState().toString());
            hosts.add(host);
        }
        return hosts;
    }

}
