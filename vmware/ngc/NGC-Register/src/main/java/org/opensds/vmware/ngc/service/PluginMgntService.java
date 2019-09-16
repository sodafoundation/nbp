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

import org.opensds.vmware.ngc.model.PluginConfigInfo;
import org.opensds.vmware.ngc.model.VCenterInfo;
import org.opensds.vmware.ngc.utils.XmlUtils;
import org.opensds.vmware.ngc.bean.ErrorCode;
import org.opensds.vmware.ngc.utils.CommonUtils;
import com.vmware.vim25.mo.ExtensionManager;
import com.vmware.vim25.mo.ServiceInstance;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.stereotype.Service;

import java.io.*;
import java.net.MalformedURLException;
import java.net.URL;
import java.rmi.RemoteException;
import java.util.Properties;

@Service
public class PluginMgntService {

    private static final Logger logger = LogManager.getLogger(PluginMgntService.class);

    private static final String PLUGININFO_PROP_FILE = "static/config/PluginInfo.properties";

    private static final String CONST_PLUGIN_NAME = "plugin.name";

    private static final String CONST_PLUGIN_KEY = "plugin.key";

    private static final String CONST_PLUGIN_COMANNY = "plugin.company.name";

    private static final String CONST_PLUGIN_SUMMARY = "plugin.summary";

    private static final String CONST_PLUGIN_EXTENSIONTYPE = "plugin.extensiontype";

    private static final String CONST_PLUGIN_VERSION = "plugin.version";

    private static final String CONST_PLUGIN_ADMIN_EMAIL = "plugin.admin.email";

    private static final String CONST_PLUGIN_THUMBPRINT = "plugin.ssl.thumbprint";

    private RegistViService regViservice = new RegistViService();


    /**
     * Plugin register process
     * @param vcInfo
     * @return ErrorCode
     */
    public ErrorCode register(VCenterInfo vcInfo) {

        ErrorCode result = CommonUtils.checkRegisterParameters(vcInfo);
        if (ErrorCode.SUCCESS.getErrodCode() != result.getErrodCode()) {
            return result;
        }

        PluginConfigInfo pluginInfo = getConfigProperty(PLUGININFO_PROP_FILE);
        if (pluginInfo == null ) {
            return ErrorCode.PLUGIN_PROPERTIES_FAILED;
        }

        try {
            URL url = new URL(CommonUtils.createVcUrl(vcInfo.getvCenterIp()));
            ServiceInstance si = new ServiceInstance(url, vcInfo.getvCenterUser(), vcInfo.getvCenterPassword(), true);
            String localHostIp = si.getSessionManager().getCurrentSession().getIpAddress();
            if (localHostIp == null || localHostIp.isEmpty()) {
                return ErrorCode.CONNECT_FAIL_GET_LOCAL_IP;
            }
            if (si.getExtensionManager().findExtension(pluginInfo.getPluginKey()) != null) {
                return ErrorCode.ALREADY_REGISTER;
            }
            result = regViservice.registerAction(si, localHostIp, pluginInfo);
            if (ErrorCode.SUCCESS.getErrodCode() != result.getErrodCode()) {
                logger.error("The plugin register failed!");
                return result;
            }
            vcInfo.setRegisterFlag("1");
            result = XmlUtils.getXml().xmlWriteVcenterInfo(vcInfo);
            if (ErrorCode.SUCCESS.getErrodCode() != result.getErrodCode()) {
                return result;
            }
            return ErrorCode.SUCCESS;
        } catch (RemoteException ex) {
            if (ex instanceof com.vmware.vim25.InvalidLogin) {
                return ErrorCode.INVALID_LOGIN;
            }
            logger.error("Register plugin error, msg is " + ex.getMessage());
            return ErrorCode.ALREADY_REGISTER;
        } catch (MalformedURLException ex) {
            logger.error("Can not create url for vCenter  " + vcInfo.getvCenterIp());
            return ErrorCode.ALREADY_REGISTER;
        } catch (Exception ex) {
            logger.error("Write the register info error, msg is " + ex.getMessage());
            return ErrorCode.ALREADY_REGISTER;
        }
    }

    /**
     * Plugin register process
     * @param vcInfo
     * @return ErrorCode
     */
    public ErrorCode unRegister(VCenterInfo vcInfo) {

        ErrorCode result = CommonUtils.checkRegisterParameters(vcInfo);
        if (ErrorCode.SUCCESS.getErrodCode() != result.getErrodCode()) {
            return result;
        }
        try {
            URL url = new URL(CommonUtils.createVcUrl(vcInfo.getvCenterIp()));
            ServiceInstance si = new ServiceInstance(url, vcInfo.getvCenterUser(), vcInfo.getvCenterPassword(), true);
            if (si.getSessionManager().getCurrentSession().getIpAddress().isEmpty()) {
                return ErrorCode.CONNECT_FAIL_GET_LOCAL_IP;
            }
            PluginConfigInfo pluginInfo = getConfigProperty(PLUGININFO_PROP_FILE);
            if (pluginInfo == null || vcInfo == null) {
                return ErrorCode.PLUGIN_PROPERTIES_FAILED;
            }
            if (si.getExtensionManager().findExtension(pluginInfo.getPluginKey()) == null) {
                return ErrorCode.NOT_ALREADY_REGISTER;
            }
            ExtensionManager extensionManager = si.getExtensionManager();
            extensionManager.unregisterExtension(pluginInfo.getPluginKey());
            result = XmlUtils.getXml().xmlWriteVcenterInfo(new VCenterInfo());
            if (ErrorCode.SUCCESS.getErrodCode() != result.getErrodCode()) {
                return result;
            }
            logger.info("Unregister the plugin success!");
            return ErrorCode.SUCCESS;

        } catch (RemoteException ex) {
            if (ex instanceof com.vmware.vim25.InvalidLogin) {
                return ErrorCode.INVALID_LOGIN;
            }
            logger.error(String.format("Unregister plugin connect error : %s.", ex.getMessage()));
            return ErrorCode.CONNECT_FAIL;
        } catch (MalformedURLException ex) {
            logger.error(String.format("Can not create url for vCenter: ", vcInfo.getvCenterIp()));
            return ErrorCode.CONNECT_FAIL;
        }
    }

    /**
     * get property form PluginInfo.properties
     * @param filePath
     * @return
     */
    private PluginConfigInfo getConfigProperty (String filePath) {
        if (filePath == null && filePath.isEmpty()) {
            logger.error("The file path of plugin property is null.");
            return null;
        }
        InputStream in = null;
        PluginConfigInfo pluginInfo = null;
        Properties property = new Properties();
        try {
            InputStream stream = getClass().getClassLoader().getResourceAsStream(filePath);
            in = new BufferedInputStream(stream);
            property.load(in);
            if (property != null & !property.isEmpty()) {
                pluginInfo = new PluginConfigInfo();
                pluginInfo.setPluginName(property.getProperty(CONST_PLUGIN_NAME));
                pluginInfo.setPluginKey(property.getProperty(CONST_PLUGIN_KEY));
                pluginInfo.setCompanyName(property.getProperty(CONST_PLUGIN_COMANNY));
                pluginInfo.setExtensiontype(property.getProperty(CONST_PLUGIN_EXTENSIONTYPE));
                pluginInfo.setVersion(property.getProperty(CONST_PLUGIN_VERSION));
                pluginInfo.setPluginSummary(property.getProperty(CONST_PLUGIN_SUMMARY));
                pluginInfo.setAdminEmail(property.getProperty(CONST_PLUGIN_ADMIN_EMAIL));
                pluginInfo.setThumbprint(property.getProperty(CONST_PLUGIN_THUMBPRINT));
            }
        } catch (IOException ex) {
            logger.error( "Read properties file exception: " + ex.getMessage());
        } finally {
            closeStream(in);
        }
        return pluginInfo;
    }

    /**
     * safety cloes stream
     * @param stream
     */
    private void closeStream (Closeable stream) {
        if (null != stream) {
            try {
                stream.close();
            } catch (IOException ex) {
                logger.error( "Close file stream exception: " + ex.getMessage());
            }
        }
    }
}
