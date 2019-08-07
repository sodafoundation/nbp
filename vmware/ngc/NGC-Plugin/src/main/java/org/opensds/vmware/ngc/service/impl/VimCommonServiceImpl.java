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

import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.expections.InactiveSessionException;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import com.vmware.vim25.*;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.models.VolumeMO;

import javax.net.ssl.HostnameVerifier;
import javax.net.ssl.HttpsURLConnection;
import javax.net.ssl.SSLSession;
import javax.xml.ws.BindingProvider;
import javax.xml.ws.handler.MessageContext;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.*;


public class VimCommonServiceImpl {

    private static final Log logger = LogFactory.getLog(VimCommonServiceImpl.class);

    protected static final String ERROR = "error";

    protected static final String OK = "ok";

    public static final VimPortType vimPort = initializeVimPort();

    static
    {
        HostnameVerifier hostNameVerifier = new HostnameVerifier() {
            @Override public boolean verify(String urlHostName, SSLSession session) {
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
            }
            catch (KeyManagementException e) {
                logger.info(e);
            }
        }
        catch (NoSuchAlgorithmException e) {
            logger.info(e);
        }
    }

    public static class TrustAllTrustManager implements javax.net.ssl.TrustManager, javax.net.ssl.X509TrustManager {
        @Override public java.security.cert.X509Certificate[] getAcceptedIssuers() {
            return null;
        }
        @Override public void checkServerTrusted(java.security.cert.X509Certificate[] certs,
                                                 String authType)
                throws java.security.cert.CertificateException {
            return;
        }
        @Override public void checkClientTrusted(java.security.cert.X509Certificate[] certs,
                                                 String authType)
                throws java.security.cert.CertificateException {
            return;
        }
    }


    private static VimPortType initializeVimPort() {
        VimService vimService = new VimService();
        return vimService.getVimPort();
    }

    protected HostConfigManager getHostConfigManager(ManagedObjectReference hostMo,
                                                     ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        Map<String, Object> properties = getMoProperties (hostMo,
                serverInfo, VimFieldsConst.PropertyNameConst.HostSystem.ConfigManager);
        return (HostConfigManager) properties.get(VimFieldsConst.PropertyNameConst.HostSystem.ConfigManager);
    }

    protected HostConfigInfo getHostConfigInfo(ManagedObjectReference hostMo,
                                               ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        Map<String, Object> properties =
                getMoProperties(hostMo, serverInfo, VimFieldsConst.PropertyNameConst.HostSystem.Config);
        return (HostConfigInfo) properties.get(VimFieldsConst.PropertyNameConst.HostSystem.Config);
    }

    public Map<String, Object> getMoProperties(ManagedObjectReference moRef,
                                               ServerInfo serverInfo,
                                               String... properties)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg, InactiveSessionException {
        List<ObjectContent> objectContents = getMoPropertyContents(moRef, serverInfo, properties);
        return getDynamicPropertiesFromObjectContents(objectContents);
    }

    protected List<ObjectContent> getMoPropertyContents(ManagedObjectReference moRef,
                                                        ServerInfo serverInfo,
                                                        String... properties)
            throws InvalidPropertyFaultMsg, RuntimeFaultFaultMsg,InactiveSessionException {

        if (moRef == null) {
            return null;
        }
        ServiceContent serviceContent = getServiceContent(serverInfo);
        PropertyFilterSpec spec = new PropertyFilterSpec();
        spec.getPropSet().add(new PropertySpec());
        if ((properties == null || properties.length == 0)) {
            spec.getPropSet().get(0).setAll(Boolean.TRUE);
        }
        else {
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

    protected ServiceContent getServiceContent(ServerInfo serverInfo)
            throws InvalidPropertyFaultMsg, InactiveSessionException, RuntimeFaultFaultMsg  {

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

    protected String getDevicePath (HostConfigInfo config, String wwn) {
        String devicePath = null;
        if (wwn == null || config ==null) {
            return devicePath;
        }
        List<ScsiLun> scsiLuns = config.getStorageDevice().getScsiLun();
        for (ScsiLun scsiLun : scsiLuns)
        {
            if (scsiLun instanceof HostScsiDisk && scsiLun.getCanonicalName().contains(wwn))
            {
                devicePath = ((HostScsiDisk) scsiLun).getDevicePath();
                break;
            }
        }
        return devicePath;
    }

    protected ManagedObjectReference createVmfsDS(ManagedObjectReference hostDatastoreSystem,
                                                  DatastoreInfo datastoreInfo,
                                                  VolumeMO volumeMO)
            throws HostConfigFaultFaultMsg, NotFoundFaultMsg, RuntimeFaultFaultMsg, DuplicateNameFaultMsg{
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
            vmfsSpec.getVmfs().setVolumeName(datastoreInfo.getDatastoreName());
            vmfsSpec.getVmfs().setMajorVersion(Integer.valueOf(datastoreInfo.getFileVersion().substring(4)));
            return vimPort.createVmfsDatastore(hostDatastoreSystem, vmfsSpec);
        }
        return null;
    }
}
