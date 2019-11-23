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

import com.vmware.vim25.*;
import org.opensds.vmware.ngc.base.TaskInfoConst;
import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.config.MoConverter;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.VolumeInfo;
import org.opensds.vmware.ngc.model.datastore.Datastore;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.model.datastore.NFSDatastore;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import org.opensds.vmware.ngc.service.DatastoreService;
import org.opensds.vmware.ngc.service.Vmservice;
import org.opensds.vmware.ngc.task.TaskExecution;
import org.opensds.vmware.ngc.task.TaskProcessor;
import org.opensds.vmware.ngc.task.createDatastore.CreateDatastoreTask;
import org.opensds.vmware.ngc.task.createDatastore.CreateLunTask;
import org.opensds.vmware.ngc.task.createDatastore.MountLunTask;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import static org.opensds.vmware.ngc.base.DatastoreTypeEnum.NFS_DATASTORE;
import static org.opensds.vmware.ngc.base.DatastoreTypeEnum.VMFS_DATASTORE;
import static org.opensds.vmware.ngc.base.TaskInfoConst.Type.TASK_DELETE_LUN;
import static org.opensds.vmware.ngc.base.TaskInfoConst.Type.TASK_EXTNED_LUN;

import java.util.*;


@Service
public class DatastoreServiceImpl extends VimCommonServiceImpl implements DatastoreService{

    private static final Log logger = LogFactory.getLog(DatastoreServiceImpl.class);

    @Autowired
    private DeviceRepository deviceRepository;

    @Autowired
    private HostServiceImpl hostService;

    @Autowired
    private Vmservice vmService;

    private static final String ERROR = "error";
    private static final String OK = "ok";

    /**
     * Create datastore
     * @param hostMos
     * @param serverInfo
     * @param datastore
     * @return
     */
    @Override
    public ResultInfo<Object> create(ManagedObjectReference[] hostMos,
                                     ServerInfo serverInfo,
                                     Datastore datastore) {
        ResultInfo resultInfo = new ResultInfo();
        Storage storage = getStorageFormDatastoreInfo(datastore, resultInfo);
        try {
            if (datastore.getType().equals(VMFS_DATASTORE.getType())) {
                createVmfsDatastore(hostMos, serverInfo, datastore, storage);
            } else if (datastore.getType().equals(NFS_DATASTORE.getType())) {
                createNfsDatastore(hostMos, serverInfo, datastore, storage);
            }
            resultInfo.setStatus(OK);
            return resultInfo;
        } catch (Exception e) {
            resultInfo.setMsg(e.getMessage());
            resultInfo.setStatus(ERROR);
            return resultInfo;
        }
    }

    /**
     * Extend datastore size
     * @param serverInfo
     * @param datastore
     * @return
     */
    @Override
    public ResultInfo<Object> extendSize(ServerInfo serverInfo,
                                         Datastore datastore) {
        ResultInfo resultInfo = new ResultInfo();
        try {
            if (datastore instanceof VMFSDatastore) {
                extendVMFSDatastore(serverInfo, datastore);
            } else if (datastore instanceof NFSDatastore) {
                //TODO: extenNFSDatastore();
            }
            resultInfo.setStatus(OK);
            return resultInfo;
        } catch (Exception e) {
            resultInfo.setMsg(e.getMessage());
            resultInfo.setStatus(ERROR);
            return resultInfo;
        }
    }

    /**
     * delete datastore
     * @param serverInfo
     * @param datastoreMo
     * @return
     */
    @Override
    public ResultInfo<Object> delete(ServerInfo serverInfo,
                                     ManagedObjectReference datastoreMo,
                                     Datastore datastore) {
        ResultInfo resultInfo = new ResultInfo();
        try {
            if (datastore instanceof VMFSDatastore) {
                deleteVMFSDatastore(serverInfo, datastore);
            } else if (datastore instanceof NFSDatastore) {
                deleteNFSDatastore(serverInfo, datastore);
            }
            resultInfo.setStatus(OK);
        } catch (Exception e) {
            resultInfo.setMsg(e.getMessage());
            resultInfo.setStatus(ERROR);
        }
        return resultInfo;
    }

    /**
     *  GET the datastore volumes info
     * @param dsMo
     * @param serverInfo
     * @return
     */
    @Override
    public ResultInfo<Object> getInfo(ManagedObjectReference dsMo,
                                      ServerInfo serverInfo) {
        logger.info(String.format(Locale.ROOT,"-----------Get daatastore(%s) info!", dsMo.getValue()));
        return vmService.getVolumesUsedByDatastore(dsMo, serverInfo);
    }

    // create vmfs datastore
    private void createVmfsDatastore (ManagedObjectReference[] hostMos,
                                      ServerInfo serverInfo,
                                      Datastore datastore,
                                      Storage storage) {
        logger.info("-----------Begin create VMFS daatastore...");
        List<TaskExecution> taskList = new ArrayList<>();
        TaskExecution createLunTask = new CreateLunTask(datastore, hostMos, serverInfo, storage);
        taskList.add(createLunTask);
        TaskExecution mountLunTask = new MountLunTask(datastore, hostMos, serverInfo, storage);
        taskList.add(mountLunTask);
        if (datastore.isCreateDatastore()) {
            TaskExecution createDSTask = new CreateDatastoreTask(datastore, hostMos, serverInfo, storage);
            taskList.add(createDSTask);
        }
        TaskProcessor.runTaskWithThread(taskList);
    }

    // create nfs datastore
    private void createNfsDatastore (ManagedObjectReference[] hostMos,
                                     ServerInfo serverInfo,
                                     Datastore datastore,
                                     Storage storage) {
        logger.info("-----------Begin create NFS daatastore...");
        logger.info("Not surpport nfs datastoreInfo!");
    }

    // extend vmfs datastore
    private void extendVMFSDatastore(ServerInfo serverInfo,
                                     Datastore datastore) throws Exception {
        logger.info("-----------Begin extend VMFS daatastore...");
        ManagedObjectReference datastoreMo = MoConverter.getMoFromUId(datastore.getId());
        List<DatastoreHostMount> datastoreHostMounts = getHostsMount(serverInfo, datastoreMo);
        ManagedObjectReference hostMo = datastoreHostMounts.get(0).getKey();
        VolumeInfo[] volumeInfos = ((VMFSDatastore)datastore).getVolumeInfos();
        Map<String, Long> extendData = null;
        for (VolumeInfo volumeInfo : volumeInfos) {
            if (volumeInfo.getExtendSize() != null && volumeInfo.getExtendSize() > 0) {
                extendData = getExtendData(serverInfo, hostMo, volumeInfo);
                extendVolume(serverInfo, volumeInfo, datastoreMo);
            }
        }
        hostService.rescanAllHba(hostMo, serverInfo);         // step 3 :extend datastore
        hostService.expandVmfsDatastoreInVolume(hostMo, serverInfo, datastoreMo, extendData);
    }

    // delete VMFS Datastore
    private void deleteVMFSDatastore(ServerInfo serverInfo,
                                     Datastore datastore) throws Exception {
        logger.info("-----------Begin delete VMFS daatastore...");
        ManagedObjectReference datastoreMo = MoConverter.getMoFromUId(datastore.getId());
        List<DatastoreHostMount> datastoreHostMounts = getHostsMount(serverInfo, datastoreMo);
        VolumeInfo[] volumeInfos = ((VMFSDatastore)datastore).getVolumeInfos();
        if (volumeInfos == null && volumeInfos.length == 0) {
            logger.error("The datastore is without volumes!");
            throw new Exception("The datastore is without volumes!");
        }
        // step 1: unmount datastore
        unmountVmfsVolume(serverInfo, datastoreMo, datastoreHostMounts);
        // step 2: remove datastore
        ManagedObjectReference usedHostMo = datastoreHostMounts.get(0).getKey();
        removeDatastore(serverInfo, datastoreMo, usedHostMo);
        // step 3: detelte volume
        deleteVolumesFromDatastore(serverInfo, usedHostMo, volumeInfos);
        // step 4: rescan HBA
        resanAllHBA(serverInfo, datastoreHostMounts);
    }

    // delete NFS Datastore
    private void deleteNFSDatastore(ServerInfo serverInfo,
                                    Datastore datastore) throws Exception {
        logger.info("-----------Begin delete NFS daatastore...");
        ManagedObjectReference datastoreMo = MoConverter.getMoFromUId(datastore.getId());
        List<DatastoreHostMount> datastoreHostMounts = getHostsMount(serverInfo, datastoreMo);
        ManagedObjectReference usedHostMo = datastoreHostMounts.get(0).getKey();
        removeDatastore(serverInfo, datastoreMo, usedHostMo); // step 1 : just remove datastore.
    }

    // @extendSize
    // step1 : get extend data info
    private Map<String, Long> getExtendData(ServerInfo serverInfo,
                                            ManagedObjectReference hostMo,
                                            VolumeInfo volumeInfo) throws Exception {
        Map<String, Long> result = new HashMap<>();
        String wwn = volumeInfo.getWwn();
        Long originalEndSector = hostService.getEndSectorNumberOfVolumeInDatastore(hostMo, wwn, serverInfo);
        Long increaseSector = volumeInfo.getExtendSize()/512L;
        Long extendEndSector =  increaseSector + originalEndSector;
        logger.info("Volume:" + wwn + "; Original EndSector: " + originalEndSector + "; extend EndSector: " + extendEndSector);
        result.put(wwn, extendEndSector);
        return result;
    }

    // @extendSize
    // step2 : extend volume
    private void extendVolume(ServerInfo serverInfo,
                              VolumeInfo volumeInfo,
                              ManagedObjectReference datastoreMo) throws Exception {
        DeviceInfo deviceInfo = deviceRepository.get(volumeInfo.getStorageId());
        Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
        if (storage == null) {
            throw new Exception("The device is not exist");
        }
        TaskInfo taskInfo = hostService.createStorageTask(datastoreMo, serverInfo, TASK_EXTNED_LUN);
        try {
            // TODO STORAGE.EXTEND(XXXX);
            hostService.changeTaskState(taskInfo, TaskInfoConst.Status.SUCCESS, String.format("Extend volume %s " +
                    "success, now LUN capcity is %s", volumeInfo.getName(), volumeInfo.getExtendSize()));
        } catch (Exception e) {
            hostService.changeTaskState(taskInfo, TaskInfoConst.Status.ERROR, String.format("Extend volume%s failed," +
                            " case : %s", volumeInfo.getName(), e.getMessage()));
            throw new Exception(String.format("Extend volume %s failed: "+ e.getMessage(), volumeInfo.getName()));
        }
    }

    // @deleteVMFSDatastore
    // step1: unmount VMFS VOLUME
    private void unmountVmfsVolume(ServerInfo serverInfo,
                                   ManagedObjectReference datastoreMo,
                                   List<DatastoreHostMount> datastoreHostMounts) throws Exception {
        for (DatastoreHostMount datastoreHostMount : datastoreHostMounts) {
            ManagedObjectReference hostMo = datastoreHostMount.getKey();
            HostConfigManager hostConfigManager = getHostConfigManager(hostMo, serverInfo);
            ManagedObjectReference hostStorageSystem = hostConfigManager.getStorageSystem();
            Map<String, Object> dpMap = getMoProperties(datastoreMo, serverInfo,
                    VimFieldsConst.PropertyNameConst.Datastore.Info);
            com.vmware.vim25.DatastoreInfo vimDatastoreInfo = (com.vmware.vim25.DatastoreInfo)
                    dpMap.get(VimFieldsConst.PropertyNameConst.Datastore.Info);
            String vmfsUuid = vimDatastoreInfo.getUrl().split("/")[5];
            logger.info("Unmount VMFS volume, vmfsUuid: " + vmfsUuid);
            vimPort.unmountVmfsVolume(hostStorageSystem, vmfsUuid);
        }
    }

    // @deleteVMFSDatastore
    // step2: remove the datastore
    private void removeDatastore(ServerInfo serverInfo,
                                 ManagedObjectReference datastoreMo,
                                 ManagedObjectReference hostMo) throws Exception {
        HostConfigManager manager = getHostConfigManager(hostMo, serverInfo);
        ManagedObjectReference hostDatastoreSystem = manager.getDatastoreSystem();
        logger.info("remove datastore!");
        vimPort.removeDatastore(hostDatastoreSystem, datastoreMo);
    }

    // @deleteVMFSDatastore
    // step3: delete the datastore
    private void deleteVolumesFromDatastore(ServerInfo serverInfo,
                                           ManagedObjectReference hostMo,
                                           VolumeInfo[] volumeInfos) throws Exception{
        for (VolumeInfo volumeInfo : volumeInfos) {
            DeviceInfo deviceInfo = deviceRepository.get(volumeInfo.getStorageId());
            Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
            if (storage == null) {
                logger.error("The device is not exist :" + deviceInfo.uid);
                continue;
            }
            TaskInfo taskInfo = hostService.createStorageTask(hostMo, serverInfo, TASK_DELETE_LUN);
            try {
                // todo : if the volumeInfo is attach and need deattach
                // todo : and delete it
                hostService.changeTaskState(taskInfo, TaskInfoConst.Status.SUCCESS, String.format("Delete volume %s " +
                        "success", volumeInfo.getName()));
            } catch (Exception e) {
                hostService.changeTaskState(taskInfo, TaskInfoConst.Status.ERROR, String.format("Delete volume%s failed," +
                        " case : %s", volumeInfo.getName(), e.getMessage()));
                throw new Exception(String.format("Extend volume %s failed: "+ e.getMessage(), volumeInfo.getName()));
            }
        }
    }

    // @deleteVMFSDatastore
    // step4: resan the all HBA in hosts
    private void resanAllHBA(ServerInfo serverInfo,
                             List<DatastoreHostMount> datastoreHostMounts) throws Exception {
        for (int i = 0; i < datastoreHostMounts.size(); i++) {
            ManagedObjectReference host =  datastoreHostMounts.get(i).getKey();
            final HostConfigManager manager1 = getHostConfigManager(host, serverInfo);
            Thread scanThread = new Thread(new Runnable() {
                @Override
                public void run() {
                    try {
                        ManagedObjectReference hostStorageSystem = manager1.getStorageSystem();
                        vimPort.rescanAllHba(hostStorageSystem);
                    } catch (Throwable e) {
                        logger.error("Rescan HBA failed!");
                    }
                }
            });
            scanThread.start();
            logger.info("Rescan HBA task started !!!");
        }
    }


    private Storage getStorageFormDatastoreInfo(Datastore datastore,
                                                ResultInfo<Object> resultInfo) {
        Storage storage = null;
        String storageID = "";
        if (datastore instanceof VMFSDatastore) {
            storageID = ((VMFSDatastore)datastore).getVolumeInfos()[0].getStorageId();
        } else if (datastore instanceof NFSDatastore) {
            storageID =  ((NFSDatastore)datastore).getStorageId();
        }
        if (storageID.isEmpty()) {
            resultInfo.setMsg("Can not get deviceId for info.");
            resultInfo.setStatus(ERROR);
            return storage;
        }
        DeviceInfo deviceInfo = deviceRepository.get(storageID);
        storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
        if (storage == null) {
            resultInfo.setMsg("The device is not exist.");
            resultInfo.setStatus(ERROR);
        }
        return storage;
    }

    private List<DatastoreHostMount> getHostsMount(ServerInfo serverInfo,
                                                   ManagedObjectReference datastoreMor) throws Exception {
        Map<String, Object> propertiesMap = getMoProperties(datastoreMor,
                serverInfo, VimFieldsConst.PropertyNameConst.Datastore.Host);
        ArrayOfDatastoreHostMount objectReferenceHosts = (ArrayOfDatastoreHostMount)
                propertiesMap.get(VimFieldsConst.PropertyNameConst.Datastore.Host);
        List<DatastoreHostMount> hostMountList = objectReferenceHosts.getDatastoreHostMount();
        return hostMountList;
    }
}
