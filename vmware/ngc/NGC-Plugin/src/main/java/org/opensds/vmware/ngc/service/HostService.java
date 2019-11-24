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

import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.model.VolumeInfo;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.opensds.vmware.ngc.entity.ResultInfo;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vim25.TaskInfo;
import com.vmware.vise.usersession.ServerInfo;

import java.util.Map;

public interface HostService {

    /**
     * Create a task in host
     *
     * @param hostMo     ManagedObjectReference mob
     * @param serverInfo Server mob
     * @param taskId     String
     * @return TaskInfo
     */
    TaskInfo createStorageTask(ManagedObjectReference hostMo, ServerInfo serverInfo, String taskId);

    /**
     * Change the task state
     *
     * @param taskInfo  task info
     * @param taskState String
     * @param message   change message
     * @return
     */
    Boolean changeTaskState(TaskInfo taskInfo, String taskState, String message);

    /**
     * Rescan the HBA in the host
     *
     * @param host       ManagedObjectReference mob
     * @param serverInfo Server mob
     * @return ResultInfo
     */
    ResultInfo<Object> rescanAllHba(ManagedObjectReference host, ServerInfo serverInfo);

    /**
     * Get datastores from the host
     *
     * @param hostMo     ManagedObjectReference mob
     * @param serverInfo Server mob
     * @return ResultInfo
     */
    ResultInfo<Object> getDatastoreList(ManagedObjectReference hostMo, ServerInfo serverInfo);

    /**
     * Get volumes from the datastore
     *
     * @param dsID       datstore id
     * @param hostMo     ManagedObjectReference mob
     * @param serverInfo ServerInfo mob
     * @return ResultInfo
     */
    ResultInfo<Object> getVolumesFromDatastore(String dsID, ManagedObjectReference hostMo, ServerInfo serverInfo);

    /**
     * Get volumes which attach to host
     * @param hostMo host mob
     * @param serverInfo service mob
     * @return list of volumes
     */
    ResultInfo<Object> getVolumeofHost(ManagedObjectReference hostMo, ServerInfo serverInfo);

    /**
     * get Host Connection state
     *
     * @param hostMo     ManagedObjectReference host
     * @param serverInfo ServerInfo serverInfo
     * @return ResultInfo
     */
    ResultInfo<Object> getHostConnectionState(ManagedObjectReference hostMo, ServerInfo serverInfo);

    /**
     * mount a volume to the host
     *
     * @param hostMos    ManagedObjectReference host list
     * @param serverInfo Server mob
     * @param storage    Storage
     * @param volumeMO   VolumeMO
     * @return ResultInfo
     */
    ResultInfo<Object> mountVolume(
            ManagedObjectReference[] hostMos, ServerInfo serverInfo, Storage storage, VolumeMO volumeMO);

    /**
     * unmount a volume form the host
     *
     * @param hostMo     host mob
     * @param serverInfo server mob
     * @param storage    storage
     * @param volumeId   volume mo
     * @return ResultInfo
     */
    ResultInfo<Object> unmountVolume(
            ManagedObjectReference hostMo, ServerInfo serverInfo, Storage storage, String volumeId);


    /**
     * mount volumes by ids
     *
     * @param hostMos    host mob
     * @param serverInfo server mob
     * @param storage    storage
     * @param ids        voulume ids
     * @return
     */
    ResultInfo<Object> mountVolumesByIds(
            ManagedObjectReference[] hostMos, ServerInfo serverInfo, Storage storage, String[] ids);

    /**
     * Convert a storage device to VMFS Datastore
     *
     * @param hostMos       ManagedObjectReference hosts
     * @param serverInfo    Server mob
     * @param volumeMO      VolumeMO
     * @param datastoreInfo VMFSDatastore
     * @return ResultInfo
     */
    ResultInfo<Object> convertVmfsDatastore(
            ManagedObjectReference[] hostMos, ServerInfo serverInfo, VolumeMO volumeMO, VMFSDatastore datastoreInfo);


    /**
     * get the luns without mount with the ESXI
     *
     * @param hostMoRef   host mob
     * @param deviceId    storage id
     * @param filterType  filter type
     * @param filterValue filter value
     * @param serverInfo  server info
     * @return num of luns
     */
    ResultInfo<Object> getMountableVolumeListCount(
            ManagedObjectReference hostMoRef,
            ServerInfo serverInfo,
            String deviceId,
            String filterType,
            String filterValue);

    /**
     * get mountable list of the ESXI
     *
     * @param hostMoRef   host mob
     * @param deviceId    device id
     * @param filterType  filterType
     * @param filterValue filterValue
     * @param start       page start
     * @param count       page count
     * @param serverInfo  sever instance
     * @return list of voumluie
     */
    ResultInfo<Object> getMountableVolumeList(
            ManagedObjectReference hostMoRef,
            ServerInfo serverInfo,
            String deviceId,
            String filterType,
            String filterValue,
            int start,
            int count);

    /**
     * Get the End Sectio Number of LUN in Datastore
     *
     * @param hostMo     ManagedObjectReference info
     * @param lunWwn     lun wwn
     * @param serverInfo server mob
     * @return EndSectorNumber
     * @throws Exception
     */
    Long getEndSectorNumberOfVolumeInDatastore(
            ManagedObjectReference hostMo, String lunWwn, ServerInfo serverInfo)
            throws Exception;

    /**
     * Expend the vmfd datostore
     *
     * @param hostMo        ManagedObjectReference hostMo
     * @param serverInfo    Server mob
     * @param datastore     Datastore mob
     * @param extendLunData expend lun
     * @return
     * @throws Exception
     */
    ManagedObjectReference expandVmfsDatastoreInVolume(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            ManagedObjectReference datastore,
            Map<String, Long> extendLunData) throws Exception;

    /**
     * Get the count of unmount the volume list from the the host
     *
     * @param hostMoRef   esix host mob
     * @param filterType  filter type string
     * @param filterValue filter type value
     * @param serverInfo  server instance
     * @return count num
     */
    ResultInfo<Object> getUnmountableVolumeListCount(
            ManagedObjectReference hostMoRef,
            ServerInfo serverInfo,
            String filterType,
            String filterValue);


    /**
     * get unmountable volume list of the ESXI
     *
     * @param hostMoRef   host mob
     * @param filterType  filterType
     * @param filterValue filterValue
     * @param start       page start
     * @param count       page count
     * @param serverInfo  sever instance
     * @return list of volumes
     */
    ResultInfo<Object> getUnmountableVolumes(
            ManagedObjectReference hostMoRef,
            ServerInfo serverInfo,
            String filterType,
            String filterValue,
            int start,
            int count);
}
