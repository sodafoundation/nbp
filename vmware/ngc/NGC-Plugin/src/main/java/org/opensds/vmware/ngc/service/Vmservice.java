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

package org.opensds.vmware.ngc.service;

import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestParam;

public interface Vmservice {

    /**
     * get the VirtualMachineDiskInfo form the vm without indepent
     * @param vmMoRef vm mob
     * @param serverInfo server mob
     * @return list of VirtualMachineDiskInfo
     */
    ResultInfo<Object> getVirtualDisks(ManagedObjectReference vmMoRef, ServerInfo serverInfo);

    /**
     * get volumes form datastore
     * @param dsMoRef ds mob
     * @param serverInfo server mob
     * @return list of volumes
     */
    ResultInfo<Object> getVolumesUsedByDatastore(ManagedObjectReference dsMoRef, ServerInfo serverInfo);

    /**
     * get the VirtualMachineDiskInfo form the vm with indenpent
     * @param vmMoRef vm mob
     * @param serverInfo server mob
     * @return list of VirtualMachineDiskInfo about raw device mappings
     */
    ResultInfo<Object> getRawDeviceMappings(ManagedObjectReference vmMoRef, ServerInfo serverInfo);

    /**
     * get the volume info belongs to the rdm
     * @param dsMoRef datastore info
     * @param volumeUuid volume uuid
     * @param serverInfo server instance
     * @return
     */
    ResultInfo<Object> getRawDeviceMappingVolumes(ManagedObjectReference dsMoRef, String volumeUuid, ServerInfo
            serverInfo);

}
