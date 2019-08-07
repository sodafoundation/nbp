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
import org.opensds.vmware.ngc.models.*;
import org.opensds.vmware.ngc.base.HostHbaEnum;
import org.opensds.vmware.ngc.base.TaskInfoConst;
import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.expections.ExpectionHandle;
import org.opensds.vmware.ngc.expections.InactiveSessionException;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import org.opensds.vmware.ngc.model.initiator.StorageHostInitiator;
import org.opensds.vmware.ngc.model.initiator.StorageHostIscsiInitiator;
import org.opensds.vmware.ngc.model.initiator.StorageHostScsiInitiator;
import org.opensds.vmware.ngc.service.HostService;
import com.vmware.vim25.*;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

import static org.opensds.vmware.ngc.base.HostHbaEnum.ISCSI;

@Service("hostserviceimpl")
public class HostServiceImpl extends VimCommonServiceImpl implements HostService{

    private static final Log logger = LogFactory.getLog(HostServiceImpl.class);

    private static final String HOSTNAME_PREFIX = "NGC_";

    private static HostServiceImpl instance = new HostServiceImpl();

    private HostServiceImpl(){}

    public static HostServiceImpl getInstance() {
        logger.info("get host instance!");
        return instance;
    }

    @Override
    public TaskInfo createStorageTask(ManagedObjectReference hostMo, ServerInfo serverInfo, String taskType) {
        logger.error("taskType : " + taskType);
        TaskInfo taskInfo = null;
        try {
            ServiceContent sc = getServiceContent(serverInfo);
            ManagedObjectReference taskManagerMo = sc.getTaskManager();
            taskInfo = vimPort.createTask(taskManagerMo, hostMo, taskType, null, false, null, null);
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

    @Override
    public Boolean changeTaskState(TaskInfo taskInfo, String taskState, String message) {
        logger.error("task change msg: " + message);
        if(taskInfo == null){
            return false;
        }
        try {
            if (message != null ) {
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

    @Override
    public ResultInfo<Object> rescanAllHba(ManagedObjectReference hostMo, ServerInfo serverInfo) {
        final ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            HostConfigManager manager = getHostConfigManager(hostMo, serverInfo);
            logger.info("Rescan all HBA started.");
            if (null != manager) {
                vimPort.rescanAllHba(manager.getStorageSystem());
            } else {
                throw new RuntimeException("HostConfigManager is null.");
            }
            logger.info("Rescan All HBA task finished.");
            return resultInfo;
        } catch (RuntimeFaultFaultMsg e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        } catch (InvalidPropertyFaultMsg e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        } catch (HostConfigFaultFaultMsg e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        } catch (InactiveSessionException e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        }
        return resultInfo;
    }

    @Override
    public ResultInfo<Object> mountVolume(
            ManagedObjectReference[] hostMos,
            ServerInfo serverInfo,
            Storage device,
            VolumeMO volumeMO) {

        logger.info(String.format("Begin mount the volume %s....", volumeMO.wwn));
        final ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            ConnectMO connectMO = null;
            for (ManagedObjectReference hostMo: hostMos) {

                List<StorageHostInitiator> hostInitiators = getHostInitiatorList(hostMo, serverInfo);

                Optional<StorageHostInitiator> iscsiInitiaor = hostInitiators.stream().filter(n ->
                        (n.getHbaType() == HostHbaEnum.ISCSI)
                ).findFirst();

                List<StorageHostInitiator> fsInitiaors = hostInitiators.stream().filter(n ->
                        (n.getHbaType() == HostHbaEnum.FC)
                ).collect(Collectors.toList());
                List<String> wwqnList = new ArrayList<>();
                fsInitiaors.forEach( n -> wwqnList.add(((StorageHostIscsiInitiator)n).getIqn()));
                connectMO = new ConnectMO(
                        createHostName(hostMo, serverInfo),
                        HOST_OS_TYPE.ESXI,
                        ((StorageHostIscsiInitiator)iscsiInitiaor.get()).getIqn(),
                        wwqnList.toArray(new String[wwqnList.size()]),
                        ATTACH_MODE.RW,
                        ATTACH_PROTOCOL.ISCSI
                );
                logger.info(String.format("Attach the volume %s to %s...",volumeMO.name, connectMO.name));
                device.attachVolume(volumeMO.id, connectMO);
            }
            resultInfo.setData(connectMO);
            resultInfo.setStatus(OK);
        } catch (RuntimeFaultFaultMsg e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        } catch (InvalidPropertyFaultMsg e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        }  catch (InactiveSessionException e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        } catch (Exception e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
        }
        return resultInfo;
    }

    @Override
    public ResultInfo<Object> convertVmfsDatastore(
            ManagedObjectReference[] hostMos,
            ServerInfo serverInfo,
            VolumeMO volumeMO,
            DatastoreInfo datastoreInfo){

        ResultInfo<Object> resultInfo =  new ResultInfo<>();
        ManagedObjectReference datastoreMo = null;
        ManagedObjectReference createdVmfsHostMo = null;
        for (ManagedObjectReference hostMo : hostMos) {
            try {
                datastoreMo = convertVmfsDatastoreFromVolume(hostMo, serverInfo, datastoreInfo, volumeMO);
            }  catch (Exception e) {
                ExpectionHandle.handleExceptions(resultInfo, e);
            }
            if (datastoreMo != null){
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

    @Override
    public ResultInfo<Object> getHostConnectionStateByHostMo(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo) {
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            Map<String, Object> propertiesMap = getMoProperties(hostMo,
                    serverInfo, VimFieldsConst.PropertyNameConst.HostSystem.Runtime);
            HostRuntimeInfo hostRuntimeInfo = (HostRuntimeInfo) propertiesMap.get(
                    VimFieldsConst.PropertyNameConst.HostSystem.Runtime);
            if (hostRuntimeInfo == null) {
                throw new RuntimeException("hostRuntimeInfo is null");
            }
            resultInfo.setData(hostRuntimeInfo.getConnectionState().name());
            return resultInfo;
        } catch (Exception e) {
            ExpectionHandle.handleExceptions(resultInfo, e);
            return resultInfo;
        }
     }

    private ManagedObjectReference convertVmfsDatastoreFromVolume(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo,
            DatastoreInfo datastoreInfo,
            VolumeMO volumeMO)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg,
            HostConfigFaultFaultMsg, NotFoundFaultMsg, DuplicateNameFaultMsg {
        logger.info(String.format("Set the vmfs format for this %s!", volumeMO.wwn));
        ManagedObjectReference result = null;
        HostConfigInfo hostConfigInfo = getHostConfigInfo(hostMo, serverInfo);
        HostConfigManager hostConfigManager = getHostConfigManager(hostMo, serverInfo);
        if (hostConfigInfo != null) {
            ManagedObjectReference hostDatastoreSystem = hostConfigManager.getDatastoreSystem();
            ManagedObjectReference hostStorageSystem = hostConfigManager.getStorageSystem();
            String devicePath = getDevicePath(hostConfigInfo, volumeMO.wwn);
            if (devicePath == null) {
                logger.info("Branch for----------------devicePath is null, begin resacnAllHba!");
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
            logger.error("Branch for----------------Set the vmfs format failed, can not find hostConfigManager");
        }
        return result;
    }

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

    private List<StorageHostInitiator> getHostInitiatorList(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        List<StorageHostInitiator> initiatorList = new ArrayList<>();
        HostConfigInfo configInfo = getHostConfigInfo(hostMo, serverInfo);
        List<HostHostBusAdapter> hostBusAdapters = configInfo.getStorageDevice().getHostBusAdapter();
        for (HostHostBusAdapter adapter : hostBusAdapters) {
            StorageHostInitiator initiator;
            if (adapter instanceof HostInternetScsiHba) {
                initiator = new StorageHostIscsiInitiator();
                HostInternetScsiHba iscsiHba = (HostInternetScsiHba) adapter;
                initiator.setHbaType(ISCSI);
                ((StorageHostIscsiInitiator) initiator).setIqn(iscsiHba.getIScsiName());
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

    private String createHostName(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
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
}
