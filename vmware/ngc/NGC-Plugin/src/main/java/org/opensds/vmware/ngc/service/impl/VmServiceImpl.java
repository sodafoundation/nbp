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
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.dao.VolumesRepository;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.expections.ExpectionHandle;
import org.opensds.vmware.ngc.expections.InactiveSessionException;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.model.VirtualMachineDiskInfo;
import org.opensds.vmware.ngc.model.VolumeInfo;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.opensds.vmware.ngc.service.Vmservice;
import org.opensds.vmware.ngc.util.CapacityUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;

@Service
public class VmServiceImpl extends VimCommonServiceImpl implements Vmservice {

    private static final Log logger = LogFactory.getLog(VimCommonServiceImpl.class);

    private static final int KBTOB = 1024;

    @Autowired
    private VolumesRepository volumesRepository;

    @Autowired
    private DeviceRepository deviceRepository;

    /**
     * get the VirtualMachineDiskInfo form the vm
     *
     * @param vmMoRef    vm mob
     * @param serverInfo server mob
     * @return list of VirtualMachineDiskInfo
     */
    @Override
    public ResultInfo<Object> getVirtualDisks(ManagedObjectReference vmMoRef, ServerInfo serverInfo) {
        logger.info("-----------Get virtual disk from VM :" + vmMoRef.getValue());
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            List<VirtualMachineDiskInfo> virDiskList = new ArrayList<>();
            for (VirtualMachineDiskInfo virtualMachineDiskInfo : getVirDiskList(vmMoRef, serverInfo)) {
                if (!virtualMachineDiskInfo.getIndept()) {  // get vm disk
                    virDiskList.add(virtualMachineDiskInfo);
                }
            }
            resultInfo.setData(virDiskList);
        } catch (Throwable ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * get volumes form datastore
     *
     * @param dsMoRef    ds mob
     * @param serverInfo server mob
     * @return list of volumes
     */
    @Override
    public ResultInfo<Object> getVolumesUsedByDatastore(ManagedObjectReference dsMoRef, ServerInfo serverInfo) {
        logger.info("-----------Get virtual Disk Storage Information:" + dsMoRef.getValue());
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            // get vwnlist
            VmfsDatastoreInfo info = getVmfsDatastoreInfo(dsMoRef, serverInfo);
            String datastoreName = getVmfsDatastoreName(dsMoRef, serverInfo);
            List<String> diskWwnList = new ArrayList<>();
            List<HostScsiDiskPartition> hostScsiDisks = info.getVmfs().getExtent();
            for (HostScsiDiskPartition hostScsiDisk : hostScsiDisks) {
                diskWwnList.add(getWWNFromLunCanonicalName(hostScsiDisk.getDiskName()));
            }
            // get host info
            List<DatastoreHostMount> hostMounts = getDatastoreHostMount(dsMoRef, serverInfo);
            if (hostMounts == null || hostMounts.size() == 0) {
                logger.error("Can not find any hosts with this datastore.");
                return resultInfo;
            }
            List<VolumeInfo> volumeInfos = new ArrayList<>();
            for (DatastoreHostMount hostMount : hostMounts) {
                volumeInfos.addAll(getVolumeListWithDisks(hostMount, serverInfo, diskWwnList, datastoreName));
            }
            resultInfo.setData(volumeInfos);
        } catch (Throwable ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * get the VirtualMachineDiskInfo form the vm with indenpent
     *
     * @param vmMoRef    vm mob
     * @param serverInfo server mob
     * @return list of VirtualMachineDiskInfo about raw device mappings
     */
    public ResultInfo<Object> getRawDeviceMappings(ManagedObjectReference vmMoRef, ServerInfo serverInfo) {
        logger.info("-----------Get raw device mapping from VM :" + vmMoRef.getValue());
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            List<VirtualMachineDiskInfo> virDiskList = new ArrayList<>();
            for (VirtualMachineDiskInfo virtualMachineDiskInfo : getVirDiskList(vmMoRef, serverInfo)) {
                if (virtualMachineDiskInfo.getIndept()) { // get RDM
                    virDiskList.add(virtualMachineDiskInfo);
                }
            }
            resultInfo.setData(virDiskList);
        } catch (Throwable ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * get the volume info belongs to the rdm
     *
     * @param dsMoRef    datastore info
     * @param volumeWWN  volume uuid
     * @param serverInfo server instance
     * @return
     */
    public ResultInfo<Object> getRawDeviceMappingVolumes(ManagedObjectReference dsMoRef, String volumeWWN, ServerInfo
            serverInfo) {
        logger.info("-----------Get volume belongs to raw device mapping from volumeid :" + volumeWWN);
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        try {
            VmfsDatastoreInfo info = getVmfsDatastoreInfo(dsMoRef, serverInfo);
            String datastoreName = getVmfsDatastoreName(dsMoRef, serverInfo);

            List<DatastoreHostMount> hostMounts = getDatastoreHostMount(dsMoRef, serverInfo);
            if (hostMounts == null || hostMounts.size() == 0) {
                logger.error("Can not find any hosts with this datastore.");
                return resultInfo;
            }
            VolumeInfo volumeInfo = new VolumeInfo();

            for (DatastoreHostMount hostMount : hostMounts) {
                Map<String, ScsiLun> scsiLunsMap = getHostScsiLunMap(hostMount.getKey(), serverInfo);
                Optional<String> opt = scsiLunsMap.keySet().stream().filter(n -> (
                        n.contains(volumeWWN))).findFirst();

                if (!opt.isPresent()) {
                    logger.error("Can not find scsiLun with " + volumeWWN);
                } else {
                    volumeInfo.updateWithScsiLun(scsiLunsMap.get(opt.get()));
                    break;
                }
            }

            DeviceInfo deviceInfo = volumesRepository.getDevicebyWWN(volumeWWN);
            Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
            if (storage != null) {
                VolumeMO volume = storage.queryVolumeByID(volumeWWN);
                volumeInfo.convertVolumeMO2Info(volume);
            }
            volumeInfo.setUsedBy(datastoreName);
            resultInfo.setData(volumeInfo);

        } catch (Throwable ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }


    // ge volume list
    private List<VolumeInfo> getVolumeListWithDisks(
            DatastoreHostMount hostMount,
            ServerInfo serverInfo,
            List<String> diskWwnList,
            String datastoreName) {
        List<VolumeInfo> volumeInfos = new ArrayList<>();
        Map<String, ScsiLun> scsiLunsMap = null;

        try {
            scsiLunsMap = getHostScsiLunMap(hostMount.getKey(), serverInfo);
        } catch (InvalidPropertyFaultMsg |RuntimeFaultFaultMsg |InactiveSessionException ex) {
            logger.error("Get scsi map error:" + ex.getMessage());
            return volumeInfos;
        }

        for (String wwn : scsiLunsMap.keySet()) {
            if (!diskWwnList.contains(wwn)) {
                continue;
            }
            ScsiLun scsiLun = scsiLunsMap.get(wwn);
            VolumeInfo volumeInfo = new VolumeInfo();
            if (scsiLun instanceof HostScsiDisk && ((HostScsiDisk) scsiLun).isLocalDisk()) {
                logger.info(String.format(Locale.ROOT, "Volume wwn %s this is localdisk", wwn));
            } else {
                try {
                    DeviceInfo deviceInfo = volumesRepository.getDevicebyWWN(wwn);
                    Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
                    if (storage != null) {
                        VolumeMO volume = storage.queryVolumeByID(wwn);
                        volumeInfo.convertVolumeMO2Info(volume);
                        volumeInfo.updateWithStorage(deviceInfo);
                        volumeInfo.updateWithPool(storage.getStoragePool(volumeInfo.getStoragePoolId()));
                    }
                } catch (Exception ex) {
                    logger.error("Get volume info error: " + ex.getMessage());
                }
            }
            volumeInfo.updateWithScsiLun(scsiLun);
            volumeInfo.setUsedBy(datastoreName);
            volumeInfos.add(volumeInfo);
        }
        return volumeInfos;
    }

    //get virdisk lists
    private List<VirtualMachineDiskInfo> getVirDiskList(
            ManagedObjectReference vmMoRef,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, NullPointerException {

        List<VirtualMachineDiskInfo> virDiskList = new ArrayList<>();
        VirtualMachineConfigInfo vmConfig = getVMConfigInfo(vmMoRef, serverInfo);
        if (vmConfig == null) {
            logger.info("This virtual machine is outline! ");
            throw new NullPointerException("This virtual machine is outline!");
        }
        VirtualHardware virHardware = vmConfig.getHardware();
        List<VirtualDevice> virDevices = virHardware.getDevice();
        if (virDevices == null || virDevices.size() == 0) {
            return virDiskList;
        }
        Map<Integer, String> virController = getVirController(virDevices);
        for (VirtualDevice virdev : virDevices) {
            VirtualMachineDiskInfo virDisk = getVirDisk(virdev, virController, serverInfo);
            if (null == virDisk) {
                continue;
            }
            virDiskList.add(virDisk);
        }
        Collections.sort(virDiskList);
        return virDiskList;
    }


    private Map<Integer, String> getVirController(List<VirtualDevice> virDevices) {
        Map<Integer, String> virController = new HashMap<Integer, String>();
        for (VirtualDevice virdev : virDevices) {
            if (virdev instanceof VirtualController) {
                VirtualController virSasControl = (VirtualController) virdev;
                String virKey = getVirControlKey(virSasControl.getDeviceInfo().getLabel());
                virController.put(virSasControl.getKey(), virKey);
            }
        }
        return virController;
    }

    private String getVirControlKey(String key) {
        StringBuffer virKey = new StringBuffer();
        if (key == null) {
            return null;
        }
        String[] temp = key.split(" ");
        virKey.append(temp[0].toLowerCase(Locale.US));
        virKey.append(temp[temp.length - 1]);
        return virKey.toString();
    }

    // get VirtualMachineDiskInfo
    private VirtualMachineDiskInfo getVirDisk(
            VirtualDevice virdev,
            Map<Integer, String> virCon,
            ServerInfo serverInfo)
            throws RuntimeFaultFaultMsg, InvalidPropertyFaultMsg, InactiveSessionException {
        VirtualMachineDiskInfo virDisk = null;
        boolean isIndependent = true;
        if (virdev == null || virCon.isEmpty()) {
            return null;
        }
        if (virdev instanceof VirtualDisk) {  //  VirtualDevice instance of VirtualDisk and VirtualController
            virDisk = new VirtualMachineDiskInfo();
            VirtualDisk virtualDisk = (VirtualDisk) virdev;
            if (virtualDisk.getCapacityInKB() == 0) {
                return null;
            }
            String virDiskName = virtualDisk.getDeviceInfo().getLabel();
            String summary = virtualDisk.getDeviceInfo().getSummary();
            String virDiskSize = getVirDiskSize(summary);
            String virDiskId = getVirDiskId(virtualDisk, virCon);
            String virDiskMode = null;
            String virFileName = null;
            ManagedObjectReference dsMoRef = null;
            String lunId = null;
            String lunWwn = null;
            VmfsDatastoreInfo vmfsDatastoreInfo = null;
            // a base data object type for information about the backing of a device in a virtual machine
            VirtualDeviceBackingInfo virBacking = virtualDisk.getBacking();
            if (virBacking instanceof VirtualDiskRawDiskMappingVer1BackingInfo) {
                isIndependent = true;
                VirtualDiskRawDiskMappingVer1BackingInfo vir1Backing = (VirtualDiskRawDiskMappingVer1BackingInfo)
                        virBacking;
                virDiskMode = getVirDiskMode(vir1Backing.getDiskMode());
                virFileName = vir1Backing.getFileName();
                lunId = vir1Backing.getLunUuid();
                dsMoRef = vir1Backing.getDatastore();
                vmfsDatastoreInfo = getVmfsDatastoreInfo(dsMoRef, serverInfo);
                List<DatastoreHostMount> hostMounts = getDatastoreHostMount(dsMoRef, serverInfo);
                lunWwn = getVirDiskLunId(hostMounts, lunId, serverInfo);
                virDisk.setCompatibilityMode(vir1Backing.getCompatibilityMode());
                virDisk.setLunUuid(lunId);
                virDisk.setLunIdentifier(lunWwn);
            } else if (virBacking instanceof VirtualDiskFlatVer2BackingInfo) {
                isIndependent = false;
                VirtualDiskFlatVer2BackingInfo vir2Backing = (VirtualDiskFlatVer2BackingInfo) virBacking;
                virDiskMode = getVirDiskMode(vir2Backing.getDiskMode());
                virFileName = vir2Backing.getFileName();
                dsMoRef = vir2Backing.getDatastore();
                vmfsDatastoreInfo = getVmfsDatastoreInfo(dsMoRef, serverInfo);
            }
            if (null == vmfsDatastoreInfo) {
                return null;
            }
            virDisk.setName(virDiskName.replace("硬盘", "Hard disk"));
            virDisk.setId(virDiskId);
            virDisk.setDiskMode(virDiskMode);
            virDisk.setDatastoreName(vmfsDatastoreInfo.getName());
            virDisk.setDiskFileName(virFileName);
            virDisk.setSize(virDiskSize);
            virDisk.setDatastoreId(getUidFromMo(dsMoRef));
            virDisk.setIndept(isIndependent);
        }
        return virDisk;
    }

    // ge vir disk size
    private String getVirDiskSize(String summary) {
        String diskSummary = summary.replace(",", "").replace("KB", "").replace(" ", "");
        long diskSize = Long.parseLong(diskSummary) * KBTOB;
        String virDiskSize = CapacityUtil.convertByteToCap(diskSize);
        return virDiskSize;
    }

    // ge vir disk id
    private String getVirDiskId(VirtualDisk virdev, Map<Integer, String> virCon) {
        if (virdev == null || virCon.isEmpty()) {
            return null;
        }
        int unitNumber = virdev.getUnitNumber();
        int controllerKey = virdev.getControllerKey();
        StringBuffer virDiskId = new StringBuffer();
        for (Map.Entry<Integer, String> m : virCon.entrySet()) {
            if (controllerKey == m.getKey()) {
                virDiskId.append(m.getValue()).append(':').append(unitNumber);
            }
        }
        return virDiskId.toString();
    }

    // get vir disk info
    private String getVirDiskMode(String mode) {
        String tempMode = mode.replace(" ", ".").replace("_", ".");
        return tempMode;
    }

    private String getVmfsDatastoreName(ManagedObjectReference datastoreMo,
                                        ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        Map<String, Object> propertiesMap = getMoProperties(datastoreMo, serverInfo,
                VimFieldsConst.PropertyNameConst.Datastore.Name);
        return (String) propertiesMap.get(VimFieldsConst.PropertyNameConst.Datastore.Name);
    }

    // get vmfs datastore info
    private VmfsDatastoreInfo getVmfsDatastoreInfo(
            ManagedObjectReference datastoreMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        VmfsDatastoreInfo vmfsDatastoreInfo;
        Map<String, Object> propertiesMap = getMoProperties(datastoreMo, serverInfo,
                VimFieldsConst.PropertyNameConst.Datastore.Info);
        if (propertiesMap.get(VimFieldsConst.PropertyNameConst.Datastore.Info) instanceof NasDatastoreInfo) {
            logger.debug("Convert NasDatastoreInfo to VmfsDatastoreInfo!");
            NasDatastoreInfo mNasDatastoreInfo = (NasDatastoreInfo) propertiesMap.get(VimFieldsConst
                    .PropertyNameConst.Datastore.Info);
            vmfsDatastoreInfo = new VmfsDatastoreInfo();
            vmfsDatastoreInfo.setName(mNasDatastoreInfo.getName());
            vmfsDatastoreInfo.setUrl(mNasDatastoreInfo.getUrl());
            vmfsDatastoreInfo.setFreeSpace(mNasDatastoreInfo.getFreeSpace());
            vmfsDatastoreInfo.setMaxFileSize(mNasDatastoreInfo.getMaxFileSize());
            vmfsDatastoreInfo.setMaxVirtualDiskCapacity(mNasDatastoreInfo.getMaxVirtualDiskCapacity());
            vmfsDatastoreInfo.setMaxMemoryFileSize(mNasDatastoreInfo.getMaxMemoryFileSize());
            vmfsDatastoreInfo.setTimestamp(mNasDatastoreInfo.getTimestamp());
            vmfsDatastoreInfo.setContainerId(mNasDatastoreInfo.getContainerId());

            HostNasVolume mHostNasVolume = mNasDatastoreInfo.getNas();
            HostFileSystemVolume mFileSystemVolume = mHostNasVolume;
            HostVmfsVolume mHostVmfsVolume = new HostVmfsVolume();
            mHostVmfsVolume.setCapacity(mFileSystemVolume.getCapacity());
            mHostVmfsVolume.setName(mFileSystemVolume.getName());
            mHostVmfsVolume.setType(mFileSystemVolume.getType());
            vmfsDatastoreInfo.setVmfs(mHostVmfsVolume);

            logger.debug("Convert VmfsDatastoreInfo to NasDatastoreInfo END!");
        } else {
            vmfsDatastoreInfo = (VmfsDatastoreInfo) propertiesMap.get(VimFieldsConst.PropertyNameConst.Datastore.Info);
        }
        return vmfsDatastoreInfo;
    }

    // get DatastoreHostMount info
    private List<DatastoreHostMount> getDatastoreHostMount(
            ManagedObjectReference datastoreMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        List<DatastoreHostMount> datastoreHostMounts;
        Map<String, Object> propertiesMap = getMoProperties(datastoreMo, serverInfo, VimFieldsConst.PropertyNameConst
                .Datastore.Host);
        datastoreHostMounts = ((ArrayOfDatastoreHostMount) propertiesMap.
                get(VimFieldsConst.PropertyNameConst.Datastore.Host)).getDatastoreHostMount();
        return datastoreHostMounts;
    }

    // get lun wwn from vm disk
    private String getVirDiskLunId(
            List<DatastoreHostMount> hostMounts,
            String uuid,
            ServerInfo serverInfo) throws RuntimeFaultFaultMsg, InvalidPropertyFaultMsg, InactiveSessionException {
        if (null == hostMounts || null == uuid) {
            return null;
        }
        ScsiLun lun = getVirLun(hostMounts, uuid, serverInfo);
        if (null == lun) {
            return null;
        }
        String[] temp = lun.getCanonicalName().split("\\.");
        String lunWwn = temp[temp.length - 1];
        return lunWwn;
    }

    //get scsi lun info from vm disk
    private ScsiLun getVirLun(
            List<DatastoreHostMount> hostMounts,
            String uuid,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        ScsiLun lun = null;
        if (hostMounts == null || hostMounts.size() == 0) {
            return null;
        }
        for (DatastoreHostMount hostMount : hostMounts) {
            ManagedObjectReference hostMo = hostMount.getKey();
            HostConfigInfo hostConfigInfo = getHostConfigInfo(hostMo, serverInfo);
            // viHost = getDatastoreHost(hostMor);
            lun = getVirDiskLun(hostConfigInfo, uuid);
            if (null != lun) {
                break;
            }
        }
        return lun;
    }

    //get the uniqued scsic lun
    private ScsiLun getVirDiskLun(HostConfigInfo hostConfigInfo, String uuid) {
        ScsiLun scsiDisk = null;
        if (hostConfigInfo == null || uuid == null) {
            return null;
        }
        List<ScsiLun> scsiLuns = hostConfigInfo.getStorageDevice().getScsiLun();
        if (null == scsiLuns || scsiLuns.size() == 0) {
            return null;
        }
        for (ScsiLun lun : scsiLuns) {
            if (lun.getUuid().equals(uuid)) {
                scsiDisk = lun;
                break;
            }
        }
        return scsiDisk;
    }
}
