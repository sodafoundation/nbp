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

package org.opensds.vmware.ngc.controller;

import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.service.Vmservice;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.ResponseBody;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestParam;

@Controller
@RequestMapping(value = "/data/vm")
public class VirtualMachineController {

    private static final Log logger = LogFactory.getLog(VirtualMachineController.class);

    @Autowired
    private Vmservice vmService;

    /**
     * get disks info from vm
     * @param vmMoRef vm mob
     * @param serverInfo server instance
     * @return list of VirtualMachineDiskInfo
     */
    @RequestMapping(value = "/virtualDisks/{vmId}", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getVirtualDisks(
            @PathVariable("vmId") ManagedObjectReference vmMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return vmService.getVirtualDisks(vmMoRef, serverInfo);
    }

    /**
     * get diskinfo from datastore
     * @param dsMoRef datastore mob
     * @param serverInfo server instance
     * @return list of volumes
     */
    @RequestMapping(value = "/virtualDiskStorageInformation/{datastoreId}", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getVirtualDiskStorageInformation(
            @PathVariable("datastoreId") ManagedObjectReference dsMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return vmService.getVolumesUsedByDatastore(dsMoRef, serverInfo);
    }

    /**
     * get raw disks info from vm
     * @param vmMoRef vm mob
     * @param serverInfo server instance
     * @return list of VirtualMachineDiskInfo about raw device mappings
     */
    @RequestMapping(value = "/rawDeviceMappings/{vmId}", method = RequestMethod.GET)
    @ResponseBody
    public  ResultInfo<Object> getRawDeviceMappings(
            @PathVariable("vmId") ManagedObjectReference vmMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return vmService.getRawDeviceMappings(vmMoRef, serverInfo);
    }

    /**
     *
     * @param dsMoRef
     * @param volumeUuid
     * @param serverInfo
     * @return
     */
    @RequestMapping(value = "/RawDeviceMappingLuns/{datastoreId}")
    @ResponseBody
    public ResultInfo<Object> getRawDeviceMappingLuns(
            @PathVariable("datastoreId") ManagedObjectReference dsMoRef,
            @RequestParam(value = "lunUuid") String volumeUuid,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return vmService.getRawDeviceMappingVolumes(dsMoRef, volumeUuid, serverInfo);
    }
}
