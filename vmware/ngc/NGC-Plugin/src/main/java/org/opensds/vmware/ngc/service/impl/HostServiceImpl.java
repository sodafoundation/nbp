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

import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.dao.VolumesRepository;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.model.StoragePoolInfo;
import org.opensds.vmware.ngc.model.VolumeInfo;
import org.opensds.vmware.ngc.model.datastore.Datastore;
import org.opensds.vmware.ngc.model.datastore.NFSDatastore;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import org.opensds.vmware.ngc.models.*;
import org.opensds.vmware.ngc.base.HostHbaEnum;
import org.opensds.vmware.ngc.base.TaskInfoConst;
import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.expections.ExpectionHandle;
import org.opensds.vmware.ngc.expections.InactiveSessionException;
import org.opensds.vmware.ngc.model.initiator.StorageHostInitiator;
import org.opensds.vmware.ngc.model.initiator.StorageHostIscsiInitiator;
import org.opensds.vmware.ngc.model.initiator.StorageHostScsiInitiator;
import org.opensds.vmware.ngc.service.HostService;
import com.vmware.vim25.*;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.util.FilterUtils;
import org.opensds.vmware.ngc.util.ListUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.*;
import java.util.stream.Collectors;

import static org.opensds.vmware.ngc.base.HostHbaEnum.ISCSI;

@Service
public class HostServiceImpl extends VimCommonServiceImpl implements HostService {

    private static final Log logger = LogFactory.getLog(HostServiceImpl.class);

    private static final String HOSTNAME_PREFIX = "NGC_";

    private static HostServiceImpl instance = new HostServiceImpl();

    @Autowired
    private VolumesRepository volumesRepository;

    @Autowired
    private DeviceRepository deviceRepository;

    // host with datastore info
    // key esxi uid
    // list of datastore
    private static Map<String, List<Datastore>> CACHE_DATASTORE_INFO = new ConcurrentHashMap<>();

    // volumes for mountable, key esxi uid, list of volumes
    private static Map<String, List<VolumeInfo>> CACHE_MOUNTABLE_VOLUME_INFO = new ConcurrentHashMap<>();

    // volums for unmountable,key esxi uid, list of volumes
    private static Map<String, List<VolumeInfo>> CACHE_UNMOUNTABLE_VOLUME_INFO = new ConcurrentHashMap<>();

    private HostServiceImpl() {
    }

    public static HostServiceImpl getInstance() {
        logger.info("get host instance!");
        return instance;
    }

    /**
     * Get the datastore list from the host
     *
     * @param hostMo     ManagedObjectReference mob
     * @param serverInfo Server mob
     * @return list of datastores
     */
    @Override
    public ResultInfo<Object> getDatastoreList(ManagedObjectReference hostMo, ServerInfo serverInfo) {
        logger.info("-----------Get the datastore list from the host!");
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        List<Datastore> datastoreList = new ArrayList<>();
        try {
            List<ManagedObjectReference> hostDatastoreMoList = getDatasroreMobList(hostMo, serverInfo);
            // exctue in concurrent
            ExecutorService executorService = Executors.newFixedThreadPool(
                    10 > hostDatastoreMoList.size() ? hostDatastoreMoList.size() : 10);
            List<Future<Datastore>> futureList = new ArrayList<>();
            for (final ManagedObjectReference datastoreMo : hostDatastoreMoList) {
                futureList.add(executorService.submit((Callable) () -> {
                    logger.info(String.format(Locale.ROOT, "Query datastore(%s) info begin in thred %s, ", datastoreMo
                            .getValue(), Thread.currentThread().getId()));
                    return getDatastoreInfoByMo(datastoreMo, hostMo, serverInfo);
                }));
            }
            executorService.shutdown();
            for (Future<Datastore> future : futureList) {
                try {
                    Datastore datastore = future.get();
                    if (datastore.getName() == null || datastore.getName().isEmpty()) {
                        continue;
                    }
                    datastoreList.add(datastore);
                } catch (Exception ex) {
                    logger.error("Get future list error : " + ex.getMessage());
                }
            }
            CACHE_DATASTORE_INFO.put(getUidFromMo(hostMo), datastoreList);      //push data into cache
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        resultInfo.setData(datastoreList);
        logger.info("-----------Get the datastore list from the host, fininshed!");
        return resultInfo;
    }

    /**
     * get the luns without mount with the ESXI
     *
     * @param hostMo      host mob
     * @param deviceId    storage id
     * @param filterType  filter type
     * @param filterValue filter value
     * @param serverInfo  server info
     * @return count of mountable volumes
     */
    @Override
    public ResultInfo<Object> getMountableVolumeListCount(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            String deviceId,
            String filterType,
            String filterValue) {
        logger.info("-----------Get the volumes count for mount with the ESXI!");
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        CACHE_MOUNTABLE_VOLUME_INFO.clear();
        try {
            List<VolumeInfo> volumeInfos = getAllMountableVolumeList(hostMo, serverInfo, deviceId, filterType,
                    filterValue);
            resultInfo.setData(volumeInfos.size());
            CACHE_MOUNTABLE_VOLUME_INFO.put(formatCahceKey(hostMo, filterType, filterValue), volumeInfos);
            logger.info("-----------Get the volumes count for mount with the ESXI finished!suze:" + volumeInfos.size());
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * get mountable list of the ESXI in per page
     *
     * @param hostMo      host mob
     * @param deviceId    device id
     * @param filterType  filterType
     * @param filterValue filterValue
     * @param start       page start
     * @param count       page count
     * @param serverInfo  sever instance
     * @return list of volumeInfos
     */
    @Override
    public ResultInfo<Object> getMountableVolumeList(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            String deviceId,
            String filterType,
            String filterValue,
            int start,
            int count) {
        logger.info("-----------Get the volumes list for mount with the ESXI!");
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        List<VolumeInfo> volumeInfos = CACHE_MOUNTABLE_VOLUME_INFO.get(formatCahceKey(hostMo, filterType, filterValue));
        if (volumeInfos != null && volumeInfos.size() > 0) {
            resultInfo.setData(ListUtil.safeSubList(volumeInfos, start, start + count));
        } else {
            String errorMsg = "Can not get the mountable volumes list in with the Esxi!";
            logger.error(errorMsg);
            resultInfo.setMsg(errorMsg);
        }
        return resultInfo;
    }

    // get key for cache
    private String formatCahceKey(ManagedObjectReference hostMO, String filterType, String filterValue) {
        return getUidFromMo(hostMO) + filterType + filterValue;
    }

    /**
     * Get the count of unmount the volume list from the the host
     *
     * @param hostMo      esix host mob
     * @param filterType  filter type string
     * @param filterValue filter type value
     * @param serverInfo  server instance
     * @return count of unmountable volumes
     */
    @Override
    public ResultInfo<Object> getUnmountableVolumeListCount(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            String filterType,
            String filterValue) {
        logger.info("-----------Get the volumes count for unmountable with the ESXI!");
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        CACHE_MOUNTABLE_VOLUME_INFO.clear();
        try {
            List<VolumeInfo> volumeInfos = getAllUnMountableVolumeList(hostMo, serverInfo);
            List<VolumeInfo> reVolumeInfos = FilterUtils.filterList(volumeInfos, filterType, filterValue);
            resultInfo.setData(reVolumeInfos.size());
            CACHE_UNMOUNTABLE_VOLUME_INFO.put(formatCahceKey(hostMo, filterType, filterValue), reVolumeInfos);
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * get unmountable volume list of the ESXI
     *
     * @param hostMo      host mob
     * @param filterType  filterType
     * @param filterValue filterValue
     * @param start       page start
     * @param count       page count
     * @param serverInfo  sever instance
     * @return list of volumes
     */
    public ResultInfo<Object> getUnmountableVolumes(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            String filterType,
            String filterValue,
            int start,
            int count) {
        logger.info("-----------Get the volumes list for unmount with the ESXI!");
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        List<VolumeInfo> volumeInfos = CACHE_UNMOUNTABLE_VOLUME_INFO.get(formatCahceKey(hostMo, filterType,
                filterValue));

        if (volumeInfos != null && volumeInfos.size() > 0) {
            resultInfo.setData(ListUtil.safeSubList(volumeInfos, start, start + count));
        } else {
            String errorMsg = "Can not get the unmountable volumes list in with the Esxi!";
            logger.error(errorMsg);
            resultInfo.setMsg(errorMsg);
        }
        return resultInfo;
    }

    /**
     * Get volumes from the datastore
     *
     * @param dsID       datstore id
     * @param hostMo     ManagedObjectReference mob
     * @param serverInfo ServerInfo mob
     * @return ResultInfo list of volumes
     */
    @Override
    public ResultInfo<Object> getVolumesFromDatastore(
            String dsID,
            ManagedObjectReference hostMo,
            ServerInfo serverInfo) {
        logger.info("-----------Get the volumes list from the datastore!");
        ResultInfo<Object> resultInfo = new ResultInfo<>();

        if (!CACHE_DATASTORE_INFO.containsKey(getUidFromMo(hostMo))) {
            getDatastoreList(hostMo, serverInfo);

            if (!CACHE_DATASTORE_INFO.containsKey(getUidFromMo(hostMo))) {
                logger.error("Host info can not find in cache!");
                return resultInfo;
            }
        }

        Optional<Datastore> optDatasore = CACHE_DATASTORE_INFO.get(getUidFromMo(hostMo)).stream().filter(n ->
                (n.getId().equals(dsID))).findFirst();
        if (!optDatasore.isPresent()) {
            getDatastoreList(hostMo, serverInfo);
        }

        optDatasore = CACHE_DATASTORE_INFO.get(getUidFromMo(hostMo)).stream().filter(n ->
                (n.getId().equals(dsID))).findFirst();
        VMFSDatastore vmfsDatastore = (VMFSDatastore) optDatasore.get();

        if (vmfsDatastore.isLocal()) {
            logger.info("It is locale disk!");
            return resultInfo;
        }

        try {
            List<VolumeInfo> volumeInfos = new ArrayList();

            for (HostScsiDiskPartition partition : vmfsDatastore.getHostScsiDiskPartitionList()) {
                logger.info("Begin query lun list from datastore, partition: " + partition.getDiskName());
                String wwn = getWWNFromLunCanonicalName(partition.getDiskName().trim());
                if (wwn == null || wwn.length() == 0) {
                    continue;
                }
                DeviceInfo deviceInfo = volumesRepository.getDevicebyWWN(wwn);
                Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
                if (storage == null) {
                    logger.info("Can not found on device!");
                }

                if (storage != null && !vmfsDatastore.isLocal()) {
                    VolumeInfo tmpVolume = new VolumeInfo();
                    tmpVolume.convertVolumeMO2Info(storage.queryVolumeByID(wwn));
                    tmpVolume.updateWithPool(storage.getStoragePool(tmpVolume.getStoragePoolId()));
                    tmpVolume.updateWithStorage(deviceInfo);
                    volumeInfos.add(tmpVolume);
                }
            }

            resultInfo.setData(volumeInfos);
            logger.info("-----------Get the volumes list from the datastore finished!");
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * Get volumes which attach to host
     *
     * @param hostMo     host mob
     * @param serverInfo service mob
     * @return list of volumes
     */
    @Override
    public ResultInfo<Object> getVolumeofHost(ManagedObjectReference hostMo, ServerInfo serverInfo) {
        logger.info(String.format(Locale.ROOT, "-----------Get the volumes list from the host(%s)!", hostMo
                .getValue()));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            List<VolumeInfo> volumeInfos = getAllUnMountableVolumeList(hostMo, serverInfo);
            resultInfo.setData(volumeInfos);
            logger.info("-----------Get the volumes end, size :" + volumeInfos.size());
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * mount a volume to exsi
     *
     * @param hostMos    ManagedObjectReference host list
     * @param serverInfo Server mob
     * @param storage    Storage
     * @param volumeMO   VolumeMO
     * @return ResultInfo
     */
    @Override
    public ResultInfo<Object> mountVolume(
            ManagedObjectReference[] hostMos,
            ServerInfo serverInfo,
            Storage storage,
            VolumeMO volumeMO) {
        logger.info(String.format(Locale.ROOT, "-----------Begin mount the volume %s....", volumeMO.id));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            ConnectMO connectMO = null;
            for (ManagedObjectReference hostMo : hostMos) {
                connectMO = constructConnectMO(hostMo, serverInfo);
                storage.attachVolume(volumeMO.id, connectMO);
            }
            resultInfo.setData(connectMO);
            resultInfo.setStatus(OK);
        } catch (RuntimeFaultFaultMsg | InvalidPropertyFaultMsg | InactiveSessionException ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * unmount a volume form the host
     *
     * @param hostMo     host mob
     * @param serverInfo server mob
     * @param storage    storage
     * @param volumeId   volume mo
     * @return ResultInfo
     */
    public ResultInfo<Object> unmountVolume(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            Storage storage,
            String volumeId) {

        logger.info(String.format(Locale.ROOT, "-----------Begin unmount the volume %s....", volumeId));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            ConnectMO connectMO = constructConnectMO(hostMo, serverInfo);
            resultInfo.setData(connectMO);
            storage.detachVolume(volumeId, connectMO);
            resultInfo.setStatus(OK);
        } catch (RuntimeFaultFaultMsg | InvalidPropertyFaultMsg | InactiveSessionException ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * mount volumes by ids
     *
     * @param hostMos    host mob
     * @param serverInfo server mob
     * @param storage    storage
     * @param ids        voulume ids
     * @return boolean mount success or failed
     */
    public ResultInfo<Object> mountVolumesByIds(
            ManagedObjectReference[] hostMos,
            ServerInfo serverInfo,
            Storage storage,
            String[] ids) {
        logger.info(String.format(Locale.ROOT, "-----------Begin mount the volumes %s....", ids.toString()));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            for (ManagedObjectReference hostMo : hostMos) {
                ConnectMO connectMO = constructConnectMO(hostMo, serverInfo);
                for (String id : ids) {
                    try {
                        storage.attachVolume(id, connectMO);
                    } catch (Exception ex) {
                        logger.error(String.format(Locale.ROOT, "Mount volume(%s) failed. Error msg : %s.", id, ex
                                .getMessage()));
                    }
                }
            }
            resultInfo.setData(true);
            resultInfo.setStatus(OK);
        } catch (RuntimeFaultFaultMsg | InvalidPropertyFaultMsg | InactiveSessionException ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * convert a storage device to vmfs datastore
     *
     * @param hostMos       ManagedObjectReference hosts
     * @param serverInfo    Server mob
     * @param volumeMO      VolumeMO
     * @param datastoreInfo VMFSDatastore
     * @return
     */
    @Override
    public ResultInfo<Object> convertVmfsDatastore(
            ManagedObjectReference[] hostMos,
            ServerInfo serverInfo,
            VolumeMO volumeMO,
            VMFSDatastore datastoreInfo) {
        logger.info(String.format(Locale.ROOT, "-----------Begin convert a storage device(%s) to vmfs datastore....",
                volumeMO.id));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        ManagedObjectReference datastoreMo = null;
        ManagedObjectReference createdVmfsHostMo = null;

        for (ManagedObjectReference hostMo : hostMos) {
            try {
                datastoreMo = convertVmfsDatastoreFromVolume(hostMo, serverInfo, datastoreInfo, volumeMO);
            } catch (Exception e) {
                ExpectionHandle.handleExceptions(resultInfo, e);
            }

            if (datastoreMo != null) {
                createdVmfsHostMo = hostMo;
                break;
            }
        }
        if (datastoreMo == null) {
            resultInfo.setStatus(ERROR);
            return resultInfo;
        }
        resultInfo = rescanOtherHost(hostMos, createdVmfsHostMo, serverInfo, volumeMO);
        return resultInfo;
    }

    /**
     * get Host connection state
     *
     * @param hostMo
     * @param serverInfo
     * @return
     */
    @Override
    public ResultInfo<Object> getHostConnectionState(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo) {
        logger.info(String.format(Locale.ROOT, "-----------Get host(%s) connection state!", hostMo.getValue()));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            Map<String, Object> propertiesMap = getMoProperties(hostMo,
                    serverInfo, VimFieldsConst.PropertyNameConst.HostSystem.Runtime);
            HostRuntimeInfo hostRuntimeInfo = (HostRuntimeInfo) propertiesMap.get(
                    VimFieldsConst.PropertyNameConst.HostSystem.Runtime);
            if (hostRuntimeInfo == null) {
                throw new RuntimeException("hostRuntimeInfo is null");
            }
            String status = hostRuntimeInfo.getConnectionState().name();
            resultInfo.setData(status);
            logger.info(String.format(Locale.ROOT, "-----------Get host connection finned, status(%s)!", status));
        } catch (RuntimeFaultFaultMsg | InvalidPropertyFaultMsg | InactiveSessionException ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        } catch (RuntimeException ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);

        }
        return resultInfo;
    }

    /**
     * @param hostMo
     * @param lunWwn
     * @param serverInfo
     * @return the number of sectors
     * @throws Exception
     */
    @Override
    public Long getEndSectorNumberOfVolumeInDatastore(
            ManagedObjectReference hostMo,
            String lunWwn,
            ServerInfo serverInfo) throws Exception {
        long result = 0L;
        HostConfigInfo hostConfigInfo = getHostConfigInfo(hostMo, serverInfo);
        HostConfigManager hostConfigManager = getHostConfigManager(hostMo, serverInfo);

        if (hostConfigManager != null) {
            ManagedObjectReference hostStorageSystem = hostConfigManager.getStorageSystem();
            String devicePath = getDevicePath(hostConfigInfo, lunWwn);
            List<String> devicePathss = new ArrayList<>();
            devicePathss.add(devicePath);
            List<HostDiskPartitionInfo> partitionInfos = vimPort.retrieveDiskPartitionInfo(hostStorageSystem,
                    devicePathss);
            for (HostDiskPartitionInfo partitionInfo : partitionInfos) {
                logger.info("in vCenter deviceName is : " + partitionInfo.getDeviceName());
                List<HostDiskPartitionAttributes> diskPartitionAttributes = partitionInfo.getSpec().getPartition();
                for (HostDiskPartitionAttributes partitionAttributes : diskPartitionAttributes) {
                    result = partitionAttributes.getEndSector();
                }
            }
        }
        return new Long(result);
    }

    /**
     * extpand the vmfs datastore info
     *
     * @param hostMo        ManagedObjectReference hostMo
     * @param serverInfo    Server mob
     * @param datastore     Datastore mob
     * @param extendLunData expend lun
     * @return
     * @throws Exception
     */
    @Override
    public ManagedObjectReference expandVmfsDatastoreInVolume(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            ManagedObjectReference datastore,
            Map<String, Long> extendLunData) throws Exception {
        if (extendLunData == null || extendLunData.size() == 0) {
            throw new RuntimeException("no Luns to expand.");
        }
        logger.info(" Begin to expandVmfsDatastore.");
        ManagedObjectReference expandDS = datastore;
        for (String wwn : extendLunData.keySet()) {
            Thread.sleep(5000);
            expandDS = expandVmfsDs(hostMo, serverInfo, expandDS, wwn, extendLunData.get(wwn));
            vimPort.refreshDatastoreStorageInfo(expandDS);
        }
        logger.info("expandVmfsDatastore result:" + expandDS);
        return expandDS;
    }

    /**
     * create Task from host
     *
     * @param objectMo   mob
     * @param serverInfo server mob
     * @param taskType   task type
     * @return taskInfo
     */
    @Override
    public TaskInfo createStorageTask(ManagedObjectReference objectMo, ServerInfo serverInfo, String taskType) {
        logger.info("Create storage Task! " + taskType);
        TaskInfo taskInfo = null;
        try {
            ServiceContent sc = getServiceContent(serverInfo);
            ManagedObjectReference taskManagerMo = sc.getTaskManager();
            taskInfo = vimPort.createTask(taskManagerMo, objectMo, taskType, null, false, null, null);
            vimPort.setTaskState(taskInfo.getTask(), TaskInfoState.RUNNING, null, null);
        } catch (RuntimeFaultFaultMsg ex) {
            logger.error(ex.getMessage());
        } catch (InvalidStateFaultMsg ex) {
            logger.error(ex.getMessage());
        } catch (Exception ex) {
            logger.error(ex.getMessage());
        }
        return taskInfo;
    }

    /**
     * change task state from host
     *
     * @param taskInfo  Task mob
     * @param taskState task state
     * @param message   msg info
     * @return boolean
     */
    @Override
    public Boolean changeTaskState(TaskInfo taskInfo, String taskState, String message) {
        if (taskInfo == null) {
            return false;
        }
        try {
            if (message != null) {
                LocalizableMessage description = new LocalizableMessage();
                description.setKey(taskInfo.getKey());
                description.setMessage(message);
                vimPort.setTaskDescription(taskInfo.getTask(), description);
            }

            if (TaskInfoState.SUCCESS.name().equalsIgnoreCase(taskState)) {
                vimPort.setTaskState(taskInfo.getTask(), TaskInfoState.SUCCESS, message, null);
                return true;
            } else {
                LocalizedMethodFault localizedMethodFault = new LocalizedMethodFault();
                localizedMethodFault.setFault(new VimFault());
                vimPort.setTaskState(taskInfo.getTask(), TaskInfoState.ERROR, null, localizedMethodFault);
                return true;
            }
        } catch (RuntimeFaultFaultMsg ex) {
            logger.error(ex.getMessage());
        } catch (InvalidStateFaultMsg ex) {
            logger.error(ex.getMessage());
        } catch (Exception ex) {
            logger.error(ex.getMessage());
        }
        return false;
    }

    /**
     * resan all HBA
     *
     * @param hostMo     host mob
     * @param serverInfo server mob
     * @return ResultInfo
     */
    @Override
    public ResultInfo<Object> rescanAllHba(ManagedObjectReference hostMo, ServerInfo serverInfo) {
        logger.info("-----------Rescan all HBA started.");
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            HostConfigManager manager = getHostConfigManager(hostMo, serverInfo);
            if (manager != null) {
                vimPort.rescanAllHba(manager.getStorageSystem());
            } else {
                throw new NullPointerException("HostConfigManager is null.");
            }
            logger.info("Rescan All HBA task finished...");
            resultInfo.setStatus(OK);
            return resultInfo;
        } catch (RuntimeFaultFaultMsg | InvalidPropertyFaultMsg e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        } catch (HostConfigFaultFaultMsg | InactiveSessionException e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        }
        return resultInfo;
    }

    // extend vmfs daatstore
    protected ManagedObjectReference expandVmfsDs(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            ManagedObjectReference datastore,
            String lunWWN,
            Long endSector) throws Exception {

        HostScsiDisk expandDisk = null;
        Map<String, ScsiLun> scsiLunsMap = getHostScsiLunMap(hostMo, serverInfo);
        for (String wwn : scsiLunsMap.keySet()) {
            ScsiLun scsiLun = scsiLunsMap.get(wwn);
            if (scsiLun instanceof HostScsiDisk) {
                HostScsiDisk disk = (HostScsiDisk) scsiLun;
                if (disk.getCanonicalName().contains(lunWWN)) {
                    expandDisk = disk;
                    break;
                }
            }
        }
        HostConfigManager hostConfigManager = getHostConfigManager(hostMo, serverInfo);
        ManagedObjectReference hostDatastoreSystem = hostConfigManager.getDatastoreSystem();
        VmfsDatastoreExpandSpec expandSpec = getVmfsDatastoreExpandSpec(hostMo, expandDisk, serverInfo, datastore,
                endSector);
        return vimPort.expandVmfsDatastore(hostDatastoreSystem, datastore, expandSpec);
    }

    // convert a volume to vmfs datastore
    private ManagedObjectReference convertVmfsDatastoreFromVolume(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            VMFSDatastore datastoreInfo,
            VolumeMO volumeMO)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg,
            HostConfigFaultFaultMsg, NotFoundFaultMsg, DuplicateNameFaultMsg {
        logger.info(String.format("Set the vmfs format for this volume(%s)!", volumeMO.wwn));
        ManagedObjectReference result = null;
        HostConfigInfo hostConfigInfo = getHostConfigInfo(hostMo, serverInfo);
        HostConfigManager hostConfigManager = getHostConfigManager(hostMo, serverInfo);

        if (hostConfigInfo != null) {
            ManagedObjectReference hostDatastoreSystem = hostConfigManager.getDatastoreSystem();
            ManagedObjectReference hostStorageSystem = hostConfigManager.getStorageSystem();
            String devicePath = getDevicePath(hostConfigInfo, volumeMO.wwn);

            if (devicePath == null) {
                logger.info("Branch for-----------devicePath is null, begin resacnAllHba!");
                vimPort.rescanAllHba(hostStorageSystem);
                hostConfigInfo = getHostConfigInfo(hostMo, serverInfo);
                devicePath = getDevicePath(hostConfigInfo, volumeMO.wwn);
            }

            if (devicePath == null) {
                String message = "Can not find the volume just recently created in host! Please check " +
                        "the host configuration, it maybe the target configuration error.";
                logger.error(message);
                throw new RuntimeException(message);
            }

            List<String> devicePathss = new ArrayList<>();
            devicePathss.add(devicePath);
            // Set the disk partition format
            logger.info("Set the disk partition fomat.");
            HostDiskPartitionInfoPartitionFormat format = HostDiskPartitionInfoPartitionFormat.UNKNOWN;
            List<HostDiskPartitionInfo> partitionInfo = vimPort.retrieveDiskPartitionInfo(hostStorageSystem,
                    devicePathss);

            if (partitionInfo == null) {
                logger.error("can not createVmfsDatastore because the partitionInfo is null");
                throw new RuntimeException("can not createVmfsDatastore because the partitionInfo is null!");
            }
            // compute disk partition info
            logger.info("Compute disk partition information.");
            vimPort.computeDiskPartitionInfo(hostConfigManager.getStorageSystem(),
                    devicePath, partitionInfo.get(0).getLayout(), format.toString());
            result = createVmfsDS(hostDatastoreSystem, datastoreInfo, volumeMO);

        } else {
            logger.error("Branch for-----------Set the vmfs format failed, can not find hostConfigManager");
        }
        return result;
    }

    // rescan other host
    private ResultInfo<Object> rescanOtherHost(
            ManagedObjectReference[] hostMos,
            ManagedObjectReference createdVmfsHostMo,
            ServerInfo serverInfo,
            VolumeMO volumeMO) {
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        for (ManagedObjectReference hostMo : hostMos) {

            if (hostMo == createdVmfsHostMo) {
                continue;
            }

            logger.info(String.format("Rescan the host %s ...", hostMo.getValue()));

            try {
                HostConfigManager hostConfigManager = getHostConfigManager(hostMo, serverInfo);
                ManagedObjectReference hostStorageSystem = hostConfigManager.getStorageSystem();
                vimPort.rescanAllHba(hostStorageSystem);
                vimPort.rescanVmfs(hostStorageSystem);
                String devicePath = getDevicePath(getHostConfigInfo(hostMo, serverInfo), volumeMO.wwn);
                if (devicePath == null) {
                    String message = "Can not find the volume just recently created in host! Please check " +
                            "the host configuration, it maybe the target configuration error.";
                    logger.info(message);
                    TaskInfo taskInfo = createStorageTask(hostMo, serverInfo,
                            TaskInfoConst.Type.TASK_CHECK_HOST_CONFIG);
                    changeTaskState(taskInfo, TaskInfoConst.Status.ERROR, message);
                }
            } catch (Exception e) {
                logger.error(String.format("Rescan the host %s falied!", hostMo.getValue()));
                ExpectionHandle.handleExceptions(resultInfo, e);
            }
        }
        resultInfo.setStatus(OK);
        return resultInfo;
    }

    // get host initiatorList
    private List<StorageHostInitiator> getHostInitiatorList(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        List<StorageHostInitiator> initiatorList = new ArrayList<>();
        HostConfigInfo configInfo = getHostConfigInfo(hostMo, serverInfo);
        List<HostHostBusAdapter> hostBusAdapters = configInfo.getStorageDevice().getHostBusAdapter();
        List<HostVirtualNic> listNics = configInfo.getNetwork().getVnic();
        String hostIp = listNics.get(0).getSpec().getIp().getIpAddress();
        logger.info(String.format("Virtual Nic IP is %s.", hostIp));

        for (HostHostBusAdapter adapter : hostBusAdapters) {
            StorageHostInitiator initiator;
            if (adapter instanceof HostInternetScsiHba) {
                initiator = new StorageHostIscsiInitiator();
                HostInternetScsiHba iscsiHba = (HostInternetScsiHba) adapter;
                initiator.setHbaType(ISCSI);
                ((StorageHostIscsiInitiator) initiator).setIqn(iscsiHba.getIScsiName());
                ((StorageHostIscsiInitiator) initiator).setIp(hostIp);
                initiatorList.add(initiator);
            } else if (adapter instanceof HostFibreChannelHba) {
                initiator = new StorageHostScsiInitiator();
                HostFibreChannelHba scsiHba = (HostFibreChannelHba) adapter;
                initiator.setHbaType(HostHbaEnum.FC);
                if (adapter instanceof HostFibreChannelOverEthernetHba) {
                    initiator.setHbaType(HostHbaEnum.FCOE);
                }
                ((StorageHostScsiInitiator) initiator).setWwpn(Long.toHexString((scsiHba.getPortWorldWideName())));
                initiatorList.add(initiator);
            } else {
                logger.info(String.format("Unsupported hba %s found on host.", adapter));
            }
        }
        return initiatorList;
    }

    // construct a ConnectMo
    private ConnectMO constructConnectMO(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        List<StorageHostInitiator> hostInitiators = getHostInitiatorList(hostMo, serverInfo);

        Optional<StorageHostInitiator> iscsiInitiaor = hostInitiators.stream().filter(n ->
                (n.getHbaType() == HostHbaEnum.ISCSI) //StorageHostIscsiInitiator
        ).findFirst();

        List<StorageHostInitiator> fsInitiaors = hostInitiators.stream().filter(n ->
                (n.getHbaType() == HostHbaEnum.FC) // StorageHostScsiInitiator
        ).collect(Collectors.toList());
        List<String> wwqnList = new ArrayList<>();
        fsInitiaors.forEach(n -> wwqnList.add(((StorageHostScsiInitiator) n).getWwpn()));

        ConnectMO connectMO = new ConnectMO(
                createHostName(hostMo, serverInfo),
                HOST_OS_TYPE.ESXI,
                ((StorageHostIscsiInitiator)iscsiInitiaor.get()).getIqn(),
                ((StorageHostIscsiInitiator)iscsiInitiaor.get()).getIp(),
                wwqnList.toArray(new String[wwqnList.size()]),
                ATTACH_MODE.RW,
                ATTACH_PROTOCOL.ISCSI
        );
        return connectMO;
    }

    // create host name
    private String createHostName(
            final ManagedObjectReference hostMo,
            final ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        Map<String, Object> properties = getMoProperties(hostMo, serverInfo,
                VimFieldsConst.PropertyNameConst.HostSystem.HardWare,
                VimFieldsConst.PropertyNameConst.HostSystem.Name);
        String hostUuid = ((HostHardwareInfo) properties.get(VimFieldsConst.PropertyNameConst.HostSystem.HardWare)).
                getSystemInfo().getUuid();
        String hostConfiguredName = (String) properties.get(VimFieldsConst.PropertyNameConst.HostSystem.Name);
        String newHostName = HOSTNAME_PREFIX + hostConfiguredName + "_" + hostUuid.substring(0, 5);
        logger.info(String.format("Create the host name is %s.", newHostName));
        return newHostName;
    }

    // get all list of volumeInfo
    private List<VolumeInfo> getAllMountableVolumeList(
            final ManagedObjectReference hostMo,
            final ServerInfo serverInfo,
            final String deviceId,
            final String filterType,
            final String filterValue) throws Exception {
        List<VolumeInfo> reVolumeList = new ArrayList<>();
        try {
            DeviceInfo deviceInfo = deviceRepository.get(deviceId);
            Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
            Map<String, VolumeInfo> allVolumeInfo = new ConcurrentHashMap<>();
            if (storage != null) {
                logger.info("Get all volumes!!");
                List<VolumeMO> volumeMOs = storage.listVolumes("Status", "available");
                logger.info("Get all volumes finished!!");
                List<VolumeMO> volumeMOList = FilterUtils.filterList(volumeMOs, filterType, filterValue);
                volumeMOList.forEach(n -> {
                    VolumeInfo volumeInfo = new VolumeInfo();
                    volumeInfo.convertVolumeMO2Info(n);
                    try {
                        volumeInfo.updateWithPool(storage.getStoragePool(n.storagePoolId));
                    } catch (Exception ex) {

                    }
                    volumeInfo.updateWithStorage(deviceInfo);
                    allVolumeInfo.put(n.wwn, volumeInfo);
                });
            }
            logger.info("Test get unmount volumes!!");
            Map<String, ScsiLun> scsiLunsMap = getHostScsiLunMap(hostMo, serverInfo);
            for (String wwn : scsiLunsMap.keySet()) {
                if (allVolumeInfo.keySet().contains(wwn)) {
                    allVolumeInfo.remove(wwn);                    // todo: 后续考虑vvol 类型等
                }
            }
            reVolumeList.addAll(allVolumeInfo.values());
            logger.info("Test get unmount volumes fininshed!!");
        } catch (Exception ex) {
            logger.error(ex.getMessage());
            throw ex;
        }
        return reVolumeList;
    }

    // get the aready mount volume list
    public List<VolumeInfo> getAllUnMountableVolumeList(
            final ManagedObjectReference hostMo,
            final ServerInfo serverInfo) throws Exception {
        List<VolumeInfo> reVolumeList = new ArrayList<>();
        try {
            Map<String, ScsiLun> scsiLunsMap = getHostScsiLunMap(hostMo, serverInfo);
            //List<VirtualMachineConfigInfo> vmConfigInfos = getVMConfigInfosOfHost(hostMo, serverInfo);
            getDatastoreList(hostMo, serverInfo);
            List<Datastore> datastoreList = CACHE_DATASTORE_INFO.get(getUidFromMo(hostMo));

            if (datastoreList == null || datastoreList.size() == 0) {
                getDatastoreList(hostMo, serverInfo);
            }

            ExecutorService executorService = Executors.newFixedThreadPool(
                    10 > scsiLunsMap.size() ? scsiLunsMap.size() : 10);
            List<Future<VolumeInfo>> futureList = new ArrayList<>();
            for (String wwn : scsiLunsMap.keySet()) {
                final ScsiLun scsiLun = scsiLunsMap.get(wwn);
                futureList.add(executorService.submit((Callable) () -> {
                    return getVolumeInfoByScsiLun(wwn, scsiLun, CACHE_DATASTORE_INFO.get(getUidFromMo(hostMo)));
                }));
            }
            executorService.shutdown();
            for (Future<VolumeInfo> future : futureList) {
                if (!future.get().getName().isEmpty()) {
                    reVolumeList.add(future.get());
                }
            }
        } catch (Exception ex) {
            logger.error(ex.getMessage());
            throw ex;
        }
        return reVolumeList;
    }

    // get volume info by scsiLun
    private VolumeInfo getVolumeInfoByScsiLun(
            final String wwn,
            final ScsiLun scsiLun,
            List<Datastore> datastoreList) throws Exception {
        logger.info(String.format(Locale.ROOT, "----- Volume(%s) begin searching in Thread %s.", wwn, Thread
                .currentThread().getId()));
        logger.info("test: scsiLun:" + scsiLun.getDisplayName() + "  datastoreList:" + datastoreList.size());

        VolumeInfo volumeInfo = new VolumeInfo();
        volumeInfo.updateWithScsiLun(scsiLun);

        if (scsiLun instanceof HostScsiDisk && ((HostScsiDisk) scsiLun).isLocalDisk()) {
            logger.info(String.format(Locale.ROOT, "Volume(%s) is localdisk.", wwn));
            return volumeInfo;
        }

        try {
            DeviceInfo deviceInfo = volumesRepository.getDevicebyWWN(wwn);
            Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);

            if (storage == null) {
                logger.error(String.format(Locale.ROOT, "Volume(%s) can not find storage in plugin", wwn));
                return volumeInfo;
            }
            VolumeMO volumeMO = storage.queryVolumeByID(wwn);
            volumeInfo.convertVolumeMO2Info(volumeMO);
            volumeInfo.updateWithStorage(deviceInfo);
            volumeInfo.updateWithPool(storage.getStoragePool(volumeInfo.getStoragePoolId()));
            decideVolumeUsedby(volumeInfo, datastoreList);
        } catch (NullPointerException ex) {
            logger.error("null pointer exection" + ex.getMessage());
        } catch (Exception ex) {
            logger.error("Exception" + ex.getMessage());
        }
        return volumeInfo;
    }


    private void decideVolumeUsedby(VolumeInfo volumeInfo, List<Datastore> datastoreList) {
        for (Datastore datastore : datastoreList) {
            if (datastore instanceof VMFSDatastore) {
                VMFSDatastore vmfsDatastore = (VMFSDatastore) datastore;
                vmfsDatastore.getHostScsiDiskPartitionList().forEach(n -> {
                    if (volumeInfo.getWwn().equals(getWWNFromLunCanonicalName(n.getDiskName()))) {
                        volumeInfo.setUsedBy(vmfsDatastore.getName());
                        volumeInfo.setUsedType("Datastore");
                    }
                });
            } else if (datastore instanceof NFSDatastore) {
                // todo : left nas
            }
        }
    }
}
