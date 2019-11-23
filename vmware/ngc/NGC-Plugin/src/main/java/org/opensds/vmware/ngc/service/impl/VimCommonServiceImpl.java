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

import org.apache.commons.lang.StringUtils;
import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.dao.VolumesRepository;
import org.opensds.vmware.ngc.expections.InactiveSessionException;
import com.vmware.vim25.*;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.model.VolumeInfo;
import org.opensds.vmware.ngc.model.datastore.Datastore;
import org.opensds.vmware.ngc.model.datastore.LocalDatastore;
import org.opensds.vmware.ngc.model.datastore.NFSDatastore;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.opensds.vmware.ngc.util.CapacityUtil;
import org.opensds.vmware.ngc.util.TimeUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import javax.net.ssl.HostnameVerifier;
import javax.net.ssl.HttpsURLConnection;
import javax.net.ssl.SSLSession;
import javax.xml.ws.BindingProvider;
import javax.xml.ws.handler.MessageContext;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.*;
import java.util.regex.Pattern;

@Service
public class VimCommonServiceImpl {

    private static final Log logger = LogFactory.getLog(VimCommonServiceImpl.class);

    @Autowired
    private VolumesRepository volumesRepository;

    @Autowired
    private DeviceRepository deviceRepository;

    protected static final String ERROR = "error";

    protected static final String OK = "ok";

    public static final VimPortType vimPort = initializeVimPort();

    static {
        HostnameVerifier hostNameVerifier = new HostnameVerifier() {
            @Override
            public boolean verify(String urlHostName, SSLSession session) {
                return true;
            }
        };
        HttpsURLConnection.setDefaultHostnameVerifier(hostNameVerifier);

        javax.net.ssl.TrustManager[] trustAllCerts = new javax.net.ssl.TrustManager[1];
        javax.net.ssl.TrustManager tm = new TrustAllTrustManager();
        trustAllCerts[0] = tm;
        javax.net.ssl.SSLContext sc = null;

        try {
            sc = javax.net.ssl.SSLContext.getInstance("SSL");
            javax.net.ssl.SSLSessionContext sslsc = sc.getServerSessionContext();
            sslsc.setSessionTimeout(0);
            try {
                sc.init(null, trustAllCerts, null);
                HttpsURLConnection.setDefaultSSLSocketFactory(sc.getSocketFactory());
            } catch (KeyManagementException e) {
                logger.info(e);
            }
        } catch (NoSuchAlgorithmException e) {
            logger.info(e);
        }
    }

    public static class TrustAllTrustManager implements javax.net.ssl.TrustManager, javax.net.ssl.X509TrustManager {
        @Override
        public java.security.cert.X509Certificate[] getAcceptedIssuers() {
            return null;
        }

        @Override
        public void checkServerTrusted(java.security.cert.X509Certificate[] certs,
                                       String authType)
                throws java.security.cert.CertificateException {
            return;
        }

        @Override
        public void checkClientTrusted(java.security.cert.X509Certificate[] certs,
                                       String authType)
                throws java.security.cert.CertificateException {
            return;
        }
    }

    /**
     * get HostConfigManager
     *
     * @param hostMo     host mob
     * @param serverInfo server info
     * @return HostConfigManager
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected HostConfigManager getHostConfigManager(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        Map<String, Object> properties = getMoProperties(hostMo,
                serverInfo, VimFieldsConst.PropertyNameConst.HostSystem.ConfigManager);
        return (HostConfigManager) properties.get(VimFieldsConst.PropertyNameConst.HostSystem.ConfigManager);
    }

    /**
     * get host config info
     *
     * @param hostMo     host mob
     * @param serverInfo server
     * @return
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected HostConfigInfo getHostConfigInfo(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        Map<String, Object> properties =
                getMoProperties(hostMo, serverInfo, VimFieldsConst.PropertyNameConst.HostSystem.Config);
        return (HostConfigInfo) properties.get(VimFieldsConst.PropertyNameConst.HostSystem.Config);
    }

    /**
     * get VM config from the esxi
     *
     * @param hostMo
     * @param serverInfo
     * @return
     * @throws InactiveSessionException
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     */
    protected List<VirtualMachineConfigInfo> getVMConfigInfosOfHost(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        List<VirtualMachineConfigInfo> vmConfigs = new ArrayList<>();
        List<ManagedObjectReference> vmMoList = getHostVmMoList(hostMo, serverInfo);
        if (vmMoList == null || 0 == vmMoList.size()) {
            logger.error("Virtual Machine Array is null or empty!");
            return vmConfigs;
        }
        for (ManagedObjectReference vmMob : vmMoList) {
            try {
                VirtualMachineConfigInfo virConfig = getVMConfigInfo(vmMob, serverInfo);
                if (virConfig == null) {
                    logger.info(vmMob.getValue() + ":this virtual machine is outline!");
                    continue;
                }
                vmConfigs.add(virConfig);
            } catch(InvalidPropertyFaultMsg | InactiveSessionException |RuntimeFaultFaultMsg e) {
                logger.error(e.getMessage());
                continue;
            }
        }
        return vmConfigs;
     }


    /**
     * get VirtualMachineConfigInfo in one vm
     *
     * @param vmMo       vm mob
     * @param serverInfo server mob
     * @return VirtualMachineConfigInfo
     * @throws InactiveSessionException
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     */
    protected VirtualMachineConfigInfo getVMConfigInfo(
            ManagedObjectReference vmMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {

        VirtualMachineConfigInfo virtualMachineConfigInfo;
        Map<String, Object> propertiesMap =
                getMoProperties(vmMo, serverInfo, VimFieldsConst.PropertyNameConst.VM.Config);
        virtualMachineConfigInfo =
                (VirtualMachineConfigInfo) propertiesMap.get(VimFieldsConst.PropertyNameConst.VM.Config);
        return virtualMachineConfigInfo;
    }

    /**
     * get properties from mob
     *
     * @param moRef      mob
     * @param serverInfo server instance
     * @param properties name of properties
     * @return map with pro name and contents
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected Map<String, Object> getMoProperties(
            ManagedObjectReference moRef,
            ServerInfo serverInfo,
            String... properties)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        List<ObjectContent> objectContents = getMoPropertyContents(moRef, serverInfo, properties);
        return getDynamicPropertiesFromObjectContents(objectContents);
    }

    /**
     * get contents of mob property
     *
     * @param moRef      mob
     * @param serverInfo server instance
     * @param properties name of properties
     * @return list of content
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected List<ObjectContent> getMoPropertyContents(
            ManagedObjectReference moRef,
            ServerInfo serverInfo,
            String... properties)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {

        if (moRef == null) {
            return null;
        }
        ServiceContent serviceContent = getServiceContent(serverInfo);
        //ServiceContent serviceContent = testVimContext.getServiceContent(); //test in local
        PropertyFilterSpec spec = new PropertyFilterSpec();
        spec.getPropSet().add(new PropertySpec());
        if ((properties == null || properties.length == 0)) {
            spec.getPropSet().get(0).setAll(Boolean.TRUE);
        } else {
            spec.getPropSet().get(0).setAll(Boolean.FALSE);
        }
        spec.getPropSet().get(0).setType(moRef.getType());
        spec.getPropSet().get(0).getPathSet().addAll(Arrays.asList(properties));
        spec.getObjectSet().add(new ObjectSpec());
        spec.getObjectSet().get(0).setObj(moRef);
        spec.getObjectSet().get(0).setSkip(Boolean.FALSE);
        List<PropertyFilterSpec> propertyFilterSpecList = new ArrayList<>(1);
        propertyFilterSpecList.add(spec);
        List<ObjectContent> objectContentList = vimPort.retrieveProperties(serviceContent.getPropertyCollector(),
                propertyFilterSpecList);
        return objectContentList;
    }

    /**
     * get service content
     *
     * @param serverInfo
     * @return serviceContent
     * @throws InvalidPropertyFaultMsg
     * @throws InactiveSessionException
     * @throws RuntimeFaultFaultMsg
     */
    protected ServiceContent getServiceContent(
            ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, InactiveSessionException, RuntimeFaultFaultMsg {

        if (serverInfo == null) {
            throw new InactiveSessionException("serverInfo is null!");
        }
        String sessionCookie = serverInfo.sessionCookie;
        String serviceUrl = serverInfo.serviceUrl;

        List<String> values = new ArrayList<>();
        values.add("vmware_soap_session=" + sessionCookie);
        Map<String, List<String>> reqHeadrs = new HashMap<>();
        reqHeadrs.put("Cookie", values);

        Map<String, Object> reqContext = ((BindingProvider) vimPort).getRequestContext();
        reqContext.put(BindingProvider.ENDPOINT_ADDRESS_PROPERTY, serviceUrl);
        reqContext.put(BindingProvider.SESSION_MAINTAIN_PROPERTY, true);
        reqContext.put(MessageContext.HTTP_REQUEST_HEADERS, reqHeadrs);
        final ManagedObjectReference svcInstanceRef = new ManagedObjectReference();
        svcInstanceRef.setType(VimFieldsConst.MoTypesConst.ServiceInstance);
        svcInstanceRef.setValue(VimFieldsConst.MoTypesConst.ServiceInstance);
        ServiceContent serviceContent = vimPort.retrieveServiceContent(svcInstanceRef);
        return serviceContent;
    }

    protected Map<String, Object> getDynamicPropertiesFromObjectContents(List<ObjectContent> contentList) {
        if (contentList != null) {
            Map<String, Object> dpMap = new HashMap<>();
            ArrayList dpList = new ArrayList();

            for (ObjectContent content : contentList) {
                dpList.addAll(content.getPropSet());
            }
            for (Object dp : dpList) {
                DynamicProperty dynamicProperty = (DynamicProperty) dp;
                dpMap.put(dynamicProperty.getName(), dynamicProperty.getVal());
            }
            return dpMap;
        }
        return null;
    }

    protected String getDevicePath(HostConfigInfo config, String wwn) {
        String devicePath = null;
        if (wwn == null || config == null) {
            return devicePath;
        }
        List<ScsiLun> scsiLuns = config.getStorageDevice().getScsiLun();
        for (ScsiLun scsiLun : scsiLuns) {
            if (scsiLun instanceof HostScsiDisk && scsiLun.getCanonicalName().contains(wwn)) {
                devicePath = ((HostScsiDisk) scsiLun).getDevicePath();
                break;
            }
        }
        return devicePath;
    }

    protected ManagedObjectReference createVmfsDS(ManagedObjectReference hostDatastoreSystem,
                                                  VMFSDatastore datastoreInfo,
                                                  VolumeMO volumeMO)
            throws HostConfigFaultFaultMsg, NotFoundFaultMsg, RuntimeFaultFaultMsg, DuplicateNameFaultMsg {
        logger.info("Create vmfs datastoreInfo form lun.");
        List<HostScsiDisk> disks = vimPort.queryAvailableDisksForVmfs(hostDatastoreSystem, null);
        List<VmfsDatastoreOption> dsOptions;
        VmfsDatastoreCreateSpec vmfsSpec;
        for (HostScsiDisk disk : disks) {
            if (!disk.getCanonicalName().contains(volumeMO.wwn)) {
                continue;
            }
            dsOptions = vimPort.queryVmfsDatastoreCreateOptions(hostDatastoreSystem, disk.getDevicePath(), null);
            vmfsSpec = (VmfsDatastoreCreateSpec) dsOptions.get(0).getSpec();
            vmfsSpec.getVmfs().setVolumeName(datastoreInfo.getName());
            vmfsSpec.getVmfs().setMajorVersion(Integer.valueOf(datastoreInfo.getVmfsVersion().substring(4)));
            return vimPort.createVmfsDatastore(hostDatastoreSystem, vmfsSpec);
        }
        return null;
    }


    protected VmfsDatastoreExpandSpec getVmfsDatastoreExpandSpec(
            ManagedObjectReference hostMo,
            HostScsiDisk disk,
            ServerInfo serverInfo,
            ManagedObjectReference datastore,
            Long endSector) throws Exception {
        HostConfigManager hostConfigManager = getHostConfigManager(hostMo, serverInfo);
        ManagedObjectReference hostDatastoreSystem = hostConfigManager.getDatastoreSystem();
        List<VmfsDatastoreOption> expandOptions = vimPort.queryVmfsDatastoreExpandOptions(hostDatastoreSystem,
                datastore);
        VmfsDatastoreExpandSpec extendSpec = pickExpandSpec(disk, expandOptions);
        if (extendSpec == null) {
            throw new Exception("The LUN no space to expand!!!");
        }
        Long sectors = endSector;
        HostScsiDiskPartition extent = extendSpec.getExtent();
        logger.info("ScsiDisk partition number: " + extent.getPartition());
        HostDiskPartitionSpec spec = extendSpec.getPartition();
        List<HostDiskPartitionAttributes> partitionAttributesList = spec.getPartition();
        for (HostDiskPartitionAttributes partitionAttributes : partitionAttributesList) {
            logger.info("start sector: " + partitionAttributes.getStartSector() + "  end sector: " +
                    partitionAttributes.getEndSector());
            if (endSector > partitionAttributes.getEndSector()) {
                sectors = partitionAttributes.getEndSector();
            }
            partitionAttributes.setEndSector(sectors);
        }
        return extendSpec;
    }

    /**
     * get wwn from disk
     *
     * @param diskName String
     * @return WWN
     */
    protected static String getWWNFromLunCanonicalName(String diskName) {
        if (StringUtils.isBlank(diskName)) {
            return "";
        }
        String diskNameHeader = "naa.";
        if (Pattern.matches("^" + diskNameHeader + "[\\w]*$", diskName.trim())) {
            return diskName.substring(diskName.indexOf(diskNameHeader) + diskNameHeader.length(), diskName.length());
        } else {
            return "";
        }
    }

    /**
     * Get datastore mob ist form hostmo
     *
     * @param hostMo     host mob
     * @param serverInfo server mob
     * @return list of ds mob
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected List<ManagedObjectReference> getDatasroreMobList(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        List<ManagedObjectReference> datastoreMoList = new ArrayList<>();
        Map<String, Object> dpMap = getMoProperties(
                hostMo, serverInfo, VimFieldsConst.PropertyNameConst.HostSystem.Datastore);
        ArrayOfManagedObjectReference arrayOfManagedObjectReference =
                (ArrayOfManagedObjectReference) dpMap.get(VimFieldsConst.PropertyNameConst.HostSystem.Datastore);
        for (ManagedObjectReference mo : arrayOfManagedObjectReference.getManagedObjectReference()) {
            datastoreMoList.add(mo);
        }
        return datastoreMoList;
    }

    /**
     * get hostFileSystemMountInfo
     *
     * @param hostMo
     * @param serverInfo
     * @return
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected List<HostFileSystemMountInfo> getHostFileSystemMountInfo(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        HostConfigInfo hostConfigInfo = getHostConfigInfo(hostMo, serverInfo);
        if (hostConfigInfo != null && hostConfigInfo.getFileSystemVolume() != null) {
            return hostConfigInfo.getFileSystemVolume().getMountInfo();
        } else {
            throw new InactiveSessionException();
        }
    }

    /**
     * detect the storage device support the hardware support the acceleration
     *
     * @param infos         list of HostFileSystemMountInfo
     * @param datastoreName
     * @return vStorageSupported :Storage device supports hardware acceleration. The ESX host will use the feature
     * to offload certain storage-related operations to the device.
     * vStorageUnknown : Initial support status value.
     * vStorageUnsupported : Storage device does not support hardware acceleration.
     * The ESX host will handle all storage-related operations.
     */
    protected String getHardWareAccSupport(List<HostFileSystemMountInfo> infos, String datastoreName) {
        String vStorageSupport = "vstorageunknown";
        for (HostFileSystemMountInfo info : infos)
            if (datastoreName.equals(info.getVolume().getName()))
                vStorageSupport = info.getVStorageSupport().toLowerCase(Locale.US);
        return vStorageSupport;
    }

    /***
     * get SCSI LUN from the host
     * @param hostMo
     * @param serverInfo
     * @return
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected Map<String, ScsiLun> getHostScsiLunMap(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {

        Map<String, ScsiLun> lunMap = new HashMap<>();
        List<ScsiLun> luns;

        HostConfigInfo hostConfigInfo = getHostConfigInfo(hostMo, serverInfo);
        luns = hostConfigInfo.getStorageDevice().getScsiLun();

        if (luns == null) {
            return null;
        }
        for (ScsiLun scsiLun : luns) {
            String wwn = getWWNFromLunCanonicalName(scsiLun.getCanonicalName().trim());
            if (!VimFieldsConst.MoTypesConst.Disk.equals(scsiLun.getDeviceType())) {
                continue;
            }
            lunMap.put(wwn, scsiLun);
        }
        return lunMap;
    }

    /**
     * get datastore info by ds mob
     *
     * @param dsMo       ds mob
     * @param hostMo     host mob
     * @param serverInfo server mob
     * @return Datastore
     * @throws InvalidPropertyFaultMsg
     * @throws RuntimeFaultFaultMsg
     * @throws InactiveSessionException
     */
    protected Datastore getDatastoreInfoByMo(
            ManagedObjectReference dsMo,
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {

        Datastore datastore;
        // convert to dif datastore
        Map<String, Object> dpMap = getMoProperties(dsMo, serverInfo, VimFieldsConst.PropertyNameConst.Datastore.Info,
                VimFieldsConst.PropertyNameConst.Datastore.Summary,
                VimFieldsConst.PropertyNameConst.Datastore.OverallStatus);

        DatastoreInfo datastoreInfo = (DatastoreInfo) dpMap.get(VimFieldsConst.PropertyNameConst.Datastore.Info);
        if (datastoreInfo instanceof LocalDatastoreInfo) {  //LocalDatastoreInfo, NasDatastoreInfo, VmfsDatastoreInfo
            datastore = fillLoaclDatastore(datastoreInfo);
        } else if (datastoreInfo instanceof VmfsDatastoreInfo) {
            datastore = fillVmfsDatastore(datastoreInfo);
        } else if (datastoreInfo instanceof NasDatastoreInfo) {
            datastore = fillNasDatastore(datastoreInfo);
        } else if (datastoreInfo instanceof VvolDatastoreInfo) {
            datastore = fillNasDatastore(datastoreInfo);
        } else {
            logger.error(String.format("Foune an unknown datastore(%s) type(%s).", datastoreInfo.getName(),
                    datastoreInfo.getClass().toString()));
            return new Datastore();
            //throw new RuntimeException("Unkown storageDatastore type");
        }
        // get datatstore id
        datastore.setId(getUidFromMo(dsMo));
        // get datastore summary info
        DatastoreSummary dsSummary = (DatastoreSummary) dpMap.get(VimFieldsConst.PropertyNameConst.Datastore.Summary);
        if (!dsSummary.isAccessible()) {
            datastore.setName(dsSummary.getName() + "(inaccessible)");
        } else {
            datastore.setName(dsSummary.getName());
        }
        datastore.setCapacity(CapacityUtil.convertByteToCap(dsSummary.getCapacity()));
        datastore.setFreeCapacity(CapacityUtil.convertByteToCap(dsSummary.getFreeSpace()));
        datastore.setCapUsage(1 - ((dsSummary.getFreeSpace()*1.0) / dsSummary.getCapacity()));
        datastore.setAccessible(dsSummary.isAccessible());
        if (datastore.getType() == null) {
            datastore.setType(dsSummary.getType());
        }
        datastore.setLastTime(TimeUtil.getUTCStringFromLong(new Date().getTime()));
        //OverallStatus
        ManagedEntityStatus dsStatus =
                (ManagedEntityStatus) dpMap.get(VimFieldsConst.PropertyNameConst.Datastore.OverallStatus);
        datastore.setOverallStatus(dsStatus.value());
        //Hardware Acceleration
        List<HostFileSystemMountInfo> hostFileSystemMountInfoList = getHostFileSystemMountInfo(hostMo, serverInfo);
        datastore.setHardWareAccSupport(getHardWareAccSupport(hostFileSystemMountInfoList, datastore.getName()));
        return datastore;
    }

    /**
     * format the uid for Mo
     *
     * @param mo ManagedObjectReference object
     * @return uid
     */
    protected String getUidFromMo(ManagedObjectReference mo) {
        return mo.getType() + ":" + mo.getValue();
    }

    /**
     * get all vm in one esxi
     *
     * @param hostMo     host mob
     * @param serverInfo server instance
     * @return list of mob
     */
    protected List<ManagedObjectReference> getHostVmMoList(
            ManagedObjectReference hostMo,
            ServerInfo serverInfo)
            throws InactiveSessionException, InvalidPropertyFaultMsg, RuntimeFaultFaultMsg {
        List<ManagedObjectReference> vmMoList = new ArrayList<>();
        Map<String, Object> propertiesMap = getMoProperties(hostMo, serverInfo, VimFieldsConst.PropertyNameConst
                .HostSystem.VM);
        ArrayOfManagedObjectReference objectReference = (ArrayOfManagedObjectReference) propertiesMap.get
                (VimFieldsConst.PropertyNameConst.HostSystem.VM);
        List<ManagedObjectReference> morList = objectReference.getManagedObjectReference();
        for (ManagedObjectReference mor : morList) {
            if ("VirtualMachine".equalsIgnoreCase(mor.getType())) {
                vmMoList.add(mor);
            }
        }
        return vmMoList;
    }

    private Datastore fillNasDatastore(DatastoreInfo datastoreInfo) {
        NFSDatastore nfsDatastore = new NFSDatastore();
        NasDatastoreInfo nasDatastoreInfo = (NasDatastoreInfo) datastoreInfo;
        nfsDatastore.setType(nasDatastoreInfo.getNas().getType());
        nfsDatastore.setLocalPath(nasDatastoreInfo.getNas().getName());
        nfsDatastore.setRemoteHost(nasDatastoreInfo.getNas().getRemoteHost());
        nfsDatastore.setRemotePath(nasDatastoreInfo.getNas().getRemotePath());
        return nfsDatastore;
    }

    private Datastore fillVmfsDatastore(DatastoreInfo datastoreInfo) {
        VMFSDatastore vmfsDatastore = new VMFSDatastore();
        VmfsDatastoreInfo vmfsDatastoreInfo = (VmfsDatastoreInfo) datastoreInfo;
        if (vmfsDatastoreInfo.getVmfs() == null) {
            return vmfsDatastore;
        }
        if (vmfsDatastoreInfo.getVmfs().isLocal() != null) {
            vmfsDatastore.setLocal(vmfsDatastoreInfo.getVmfs().isLocal());
        }
        if (vmfsDatastoreInfo.getVmfs().getExtent() != null) {
            vmfsDatastore.getHostScsiDiskPartitionList().addAll(vmfsDatastoreInfo.getVmfs().getExtent());
        }
        if (vmfsDatastoreInfo.getVmfs().getUuid() != null) {
            vmfsDatastore.setUuid(vmfsDatastoreInfo.getVmfs().getUuid());
        }
        if (vmfsDatastoreInfo.getVmfs().getType() != null && vmfsDatastoreInfo.getVmfs().getMajorVersion() != 0) {
            vmfsDatastore.setType(vmfsDatastoreInfo.getVmfs().getType() + " " +
                    vmfsDatastoreInfo.getVmfs().getMajorVersion());
        }
        return vmfsDatastore;
    }


    private Datastore fillLoaclDatastore(DatastoreInfo datastoreInfo) {
        LocalDatastore localDatastore = new LocalDatastore();
        localDatastore.setPath(((LocalDatastoreInfo) datastoreInfo).getPath());
        return localDatastore;
    }

    private VmfsDatastoreExpandSpec pickExpandSpec(HostScsiDisk disk, List<VmfsDatastoreOption> expandOptions) {
        if ((expandOptions == null) || expandOptions.isEmpty()) {
            return null;
        }
        for (VmfsDatastoreOption option : expandOptions) {
            VmfsDatastoreExpandSpec spec = (VmfsDatastoreExpandSpec) option.getSpec();
            String diskName = spec.getExtent().getDiskName();
            if (diskName.equals(disk.getCanonicalName())) {
                return spec;
            }
        }
        return null;
    }

    private static VimPortType initializeVimPort() {
        //return testVimContext.get_vimPort();  //test in local
        VimService vimService = new VimService();
        return vimService.getVimPort();
    }
}
