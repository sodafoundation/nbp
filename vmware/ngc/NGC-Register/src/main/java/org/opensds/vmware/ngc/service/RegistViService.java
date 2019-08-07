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

import org.opensds.vmware.ngc.bean.ErrorCode;
import org.opensds.vmware.ngc.utils.CommonUtils;
import org.opensds.vmware.ngc.utils.XmlUtils;
import org.opensds.vmware.ngc.model.AuthInfo;
import org.opensds.vmware.ngc.model.EventsInfo;
import org.opensds.vmware.ngc.model.PluginConfigInfo;
import org.opensds.vmware.ngc.model.TaskInfo;
import com.vmware.vim25.*;
import com.vmware.vim25.mo.ExtensionManager;
import com.vmware.vim25.mo.ServiceInstance;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.rmi.RemoteException;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.List;


public class RegistViService {

    private static final Logger logger = LogManager.getLogger(RegistViService.class);

    /**
     * register action
     * @param si
     * @param hostip
     * @return
     */
    public ErrorCode registerAction (ServiceInstance si, String hostip, PluginConfigInfo pluginInfo) {
        String local = localeEdit(si.getSessionManager().getDefaultLocale());
        if (local == null) {
            return ErrorCode.CONNECT_FAIL;
        }
        if (!(local.equals("en") || local.equals("zh"))) {
            return ErrorCode.UNSUPPORT_LOCALSES;
        }
        ExtensionManager extensionManager = si.getExtensionManager();
        Extension extension = createExtension(hostip, local, pluginInfo);
        try {
            extensionManager.registerExtension(extension);
            return ErrorCode.SUCCESS;
        } catch (RuntimeFault ex) {
            logger.error("Register action runtimefault error: " + ex.getMessage());
            return ErrorCode.CONNECT_FAIL;
        } catch (RemoteException ex) {
            logger.error("Register action remoteException error: " + ex.getMessage());
            return ErrorCode.CONNECT_FAIL;
        } catch (NullPointerException ex) {
            logger.error("Register action nullPointer error: " + ex.getMessage());
            return ErrorCode.CONNECT_FAIL;
        }
    }

    /**
     * create Extension in vcetner server
     * @param hostip
     * @param local
     * @return
     */
    private Extension createExtension(String hostip, String local, PluginConfigInfo pluginConfigInfo) {
        // description of plugin
        Description description = new Description();
        description.setLabel(pluginConfigInfo.getPluginName());
        description.setSummary(pluginConfigInfo.getPluginSummary());

        // ExtensionServer info
        ExtensionServerInfo extensionServerInfo = new ExtensionServerInfo();
        extensionServerInfo.setDescription(description);
        extensionServerInfo.setCompany(pluginConfigInfo.getCompanyName());
        extensionServerInfo.setAdminEmail(new String[] {pluginConfigInfo.getCompanyName()});
        //
        if (pluginConfigInfo.getThumbprint() == null || pluginConfigInfo.getThumbprint().equals("")) {
            extensionServerInfo.setType("HTTP");
        } else {
            extensionServerInfo.setType("HTTPS");
            extensionServerInfo.setServerThumbprint(pluginConfigInfo.getThumbprint());
        }
        extensionServerInfo.setUrl(CommonUtils.createRigesterUrl(hostip));

        // ExtensinServer client info
        ExtensionClientInfo extensionClientInfo = new ExtensionClientInfo();
        extensionClientInfo.setVersion(pluginConfigInfo.getVersion());
        extensionClientInfo.setDescription(description);
        extensionClientInfo.setCompany(pluginConfigInfo.getCompanyName());
        extensionClientInfo.setType(pluginConfigInfo.getExtensiontype());
        extensionClientInfo.setUrl(CommonUtils.createRigesterUrl(hostip));

        // events info of plugins
        List<EventsInfo> eventsInfoList = XmlUtils.getXml().xmlReadEventTypeInfo(local);

        // events info of plugin
        List<TaskInfo> taskInfoList = XmlUtils.getXml().xmlReadTaskInfo(local);

        // authority of plugin
        List<AuthInfo> authInfoList = XmlUtils.getXml().xmlReadAuthInfo(local);

        if (eventsInfoList == null || taskInfoList == null || authInfoList == null) {
            logger.error(":The eventsInfoList or taskInfoList or authInfoList is none!");
            return new Extension();
        }
        List<ExtensionEventTypeInfo> exTyInfoList = new ArrayList<ExtensionEventTypeInfo>();
        List<ExtensionTaskTypeInfo> exTkInfoList = new ArrayList<ExtensionTaskTypeInfo>();
        List<ExtensionPrivilegeInfo> exPrInfoList = new ArrayList<ExtensionPrivilegeInfo>();

        // insert events in ExtensionEventTypeInfo array
        for (EventsInfo eveInfo : eventsInfoList) {
            ExtensionEventTypeInfo extensionEventTypeInfo = new ExtensionEventTypeInfo();
            extensionEventTypeInfo.setEventID(eveInfo.getEventTypeId());
            extensionEventTypeInfo.setEventTypeSchema(eveInfo.getEventTypeSchema());
            exTyInfoList.add(extensionEventTypeInfo);
        }
        int size = eventsInfoList.size();
        ExtensionEventTypeInfo[] exInfo = (ExtensionEventTypeInfo[]) exTyInfoList.toArray(new ExtensionEventTypeInfo[size]);

        // insert tasks in ExtensionTaskTypeInfo array
        for (TaskInfo taskInfo : taskInfoList) {
            ExtensionTaskTypeInfo extensionTaskTypeInfo = new ExtensionTaskTypeInfo();
            extensionTaskTypeInfo.setTaskID(taskInfo.getTaskId());
            exTkInfoList.add(extensionTaskTypeInfo);
        }
        size = exTkInfoList.size();
        ExtensionTaskTypeInfo[] extInfo = (ExtensionTaskTypeInfo[]) exTkInfoList.toArray(new ExtensionTaskTypeInfo[size]);

        // insert authInfos in ExtensionTaskTypeInfo array
        for (AuthInfo authInfo : authInfoList) {
            ExtensionPrivilegeInfo extensionPrivilegeInfo = new ExtensionPrivilegeInfo();
            extensionPrivilegeInfo.setPrivGroupName(authInfo.getAuthName());
            extensionPrivilegeInfo.setPrivID(authInfo.getAuthId());
            exPrInfoList.add(extensionPrivilegeInfo);
        }
        int psize = exPrInfoList.size();
        ExtensionPrivilegeInfo[] expInfo = (ExtensionPrivilegeInfo[]) exPrInfoList.toArray(new ExtensionPrivilegeInfo[psize]);

        // insert resource
        ExtensionResourceInfo[] exResourceInfo = createExtensionResource();
        if (exResourceInfo == null) {
            logger.error("The extesionResource of storage plugin is null!");
            return new Extension();
        }

        // make extenseion
        Extension vextension = new Extension();
        vextension.setDescription(description);
        vextension.setKey(pluginConfigInfo.getPluginKey());
        vextension.setVersion(pluginConfigInfo.getVersion());
        vextension.setSubjectName(pluginConfigInfo.getPluginName());
        vextension.setCompany(pluginConfigInfo.getCompanyName());

        vextension.setServer(new ExtensionServerInfo[] {extensionServerInfo});
        vextension.setClient(new ExtensionClientInfo[] {extensionClientInfo});
        vextension.setLastHeartbeatTime(Calendar.getInstance());
        vextension.setEventList(exInfo);
        vextension.setTaskList(extInfo);
        vextension.setResourceList(exResourceInfo);
        vextension.setPrivilegeList(expInfo);
        return vextension;
    }

    /**
     * get event and task extension resource from xml
     * @return
     */
    private ExtensionResourceInfo[] createExtensionResource() {
        //get all events
        List<EventsInfo> eventInfoList = XmlUtils.getXml().xmlReadAllEventInfo();
        //get all tasks
        List<TaskInfo> taskInfoList = XmlUtils.getXml().xmlReadAllTaskInfo();

        if (null == eventInfoList || null == taskInfoList) {
            logger.error("the eventList or taskList is null!");
            return null;
        }
        List<KeyValue> keyVaListEn = new ArrayList<>();
        List<KeyValue> keyVaListZh = new ArrayList<>();
        List<KeyValue> keyTaskListEn = new ArrayList<>();
        List<KeyValue> keyTaskListZh = new ArrayList<>();

        //create value list of event
        for (EventsInfo eveInfo : eventInfoList) {
            if (eveInfo.getEventLocal().equals("en")) {
                KeyValue keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".category");
                keyValue.setValue("info");
                keyVaListEn.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".fullFormat");
                keyValue.setValue("{message}");
                keyVaListEn.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".description");
                keyValue.setValue(eveInfo.getEventName());
                keyVaListEn.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".formatOnHost");
                keyValue.setValue("{message}");
                keyVaListEn.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId()
                        + ".formatOnDatacenter");
                keyValue.setValue("{message}");
                keyVaListEn.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".formatOnCluster");
                keyValue.setValue("{message}");
                keyVaListEn.add(keyValue);
            } else {
                KeyValue keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".category");
                keyValue.setValue("\u4fe1\u606f");
                keyVaListZh.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".fullFormat");
                keyValue.setValue("{message}");
                keyVaListZh.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".description");
                keyValue.setValue(eveInfo.getEventName());
                keyVaListZh.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".formatOnHost");
                keyValue.setValue("{message}");
                keyVaListZh.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId()
                        + ".formatOnDatacenter");
                keyValue.setValue("{message}");
                keyVaListZh.add(keyValue);
                keyValue = new KeyValue();
                keyValue.setKey(eveInfo.getEventTypeId() + ".formatOnCluster");
                keyValue.setValue("{message}");
                keyVaListZh.add(keyValue);
            }
        }
        int keyValEnSize = keyVaListEn.size();
        int keyValZhSize = keyVaListZh.size();
        KeyValue[] keyValEn = keyVaListEn.toArray(new KeyValue[keyValEnSize]);
        KeyValue[] keyValZh = keyVaListZh.toArray(new KeyValue[keyValZhSize]);

        //create value list of task
        for (TaskInfo taskInfo : taskInfoList) {
            KeyValue keyValue = new KeyValue();
            keyValue.setKey(taskInfo.getTaskLabel());
            keyValue.setValue(taskInfo.getTaskLabelValue());
            KeyValue keyValueSummary = new KeyValue();
            keyValueSummary.setKey(taskInfo.getTaskSummary());
            keyValueSummary.setValue(taskInfo.getTaskSummaryValue());
            if (taskInfo.getTasklocal().equals("en")) {
                keyTaskListEn.add(keyValue);
                keyTaskListEn.add(keyValueSummary);
            } else {
                keyTaskListZh.add(keyValue);
                keyTaskListZh.add(keyValueSummary);
            }
        }
        int keyTaskEnSize = keyTaskListEn.size();
        int keyTaskZhSize = keyTaskListZh.size();
        KeyValue[] keyTaskEn = keyTaskListEn.toArray(new KeyValue[keyTaskEnSize]);
        KeyValue[] keyTaskZh =  keyTaskListZh.toArray(new KeyValue[keyTaskZhSize]);

        List<ExtensionResourceInfo> extensionResourceInfoList = new ArrayList<ExtensionResourceInfo>();
        ExtensionResourceInfo extensionresourceinfo = new ExtensionResourceInfo();
        extensionresourceinfo.setModule("event");
        extensionresourceinfo.setLocale("en");
        extensionresourceinfo.setData(keyValEn);
        extensionResourceInfoList.add(extensionresourceinfo);
        extensionresourceinfo = new ExtensionResourceInfo();
        extensionresourceinfo.setModule("event");
        extensionresourceinfo.setLocale("zh_CN");
        extensionresourceinfo.setData(keyValZh);
        extensionResourceInfoList.add(extensionresourceinfo);
        extensionresourceinfo = new ExtensionResourceInfo();
        extensionresourceinfo.setModule("task");
        extensionresourceinfo.setLocale("en");
        extensionresourceinfo.setData(keyTaskEn);
        extensionResourceInfoList.add(extensionresourceinfo);
        extensionresourceinfo = new ExtensionResourceInfo();
        extensionresourceinfo.setModule("task");
        extensionresourceinfo.setLocale("zh_CN");
        extensionresourceinfo.setData(keyTaskZh);
        extensionResourceInfoList.add(extensionresourceinfo);
        List<ExtensionResourceInfo> permissionExResource = createPerExtensionResource();
        extensionResourceInfoList.addAll(permissionExResource);
        int exResourceInfoSize = extensionResourceInfoList.size();
        ExtensionResourceInfo[] exResourceInfo = extensionResourceInfoList.toArray(new ExtensionResourceInfo[exResourceInfoSize]);
        return exResourceInfo;
    }

    /**
     * get auth extension resouce info from xml
     * @return
     */
    private List<ExtensionResourceInfo> createPerExtensionResource() {
        List<AuthInfo> authInfoList = XmlUtils.getXml().xmlReadAllAuthInfo();
        List<KeyValue> keyAuthListEn = new ArrayList<>();
        List<KeyValue> keyAuthListZh = new ArrayList<>();
        for (AuthInfo authInfo : authInfoList) {
            KeyValue keyValue = new KeyValue();
            keyValue.setKey(authInfo.getAuthIdLabel());
            keyValue.setValue(authInfo.getAuthIdLabelValue());

            KeyValue keyValueSummary = new KeyValue();
            keyValueSummary.setKey(authInfo.getAuthIdSummary());
            keyValueSummary.setValue(authInfo.getAuthIdSummaryValue());

            KeyValue keyValueAuthNaLabel = new KeyValue();
            keyValueAuthNaLabel.setKey(authInfo.getAuthNaLabel());
            keyValueAuthNaLabel.setValue(authInfo.getAuthNaLabelValue());

            KeyValue keyValueAuthNaSummary = new KeyValue();
            keyValueAuthNaSummary.setKey(authInfo.getAuthNaSummary());
            keyValueAuthNaSummary.setValue(authInfo.getAuthNaSummaryValue());

            if (authInfo.getAuthLocal().equals("en")) {
                keyAuthListEn.add(keyValue);
                keyAuthListEn.add(keyValueSummary);
                keyAuthListEn.add(keyValueAuthNaLabel);
                keyAuthListEn.add(keyValueAuthNaSummary);
            } else {
                keyAuthListZh.add(keyValue);
                keyAuthListZh.add(keyValueSummary);
                keyAuthListZh.add(keyValueAuthNaLabel);
                keyAuthListZh.add(keyValueAuthNaSummary);
            }
        }
        int keyAuthEnSize = keyAuthListEn.size();
        int keyAuthZhSize = keyAuthListZh.size();
        KeyValue[] keyAuthEn = keyAuthListEn.toArray(new KeyValue[keyAuthEnSize]);
        KeyValue[] keyAuthZh = keyAuthListZh.toArray(new KeyValue[keyAuthZhSize]);
        List<ExtensionResourceInfo> extensionResourceInfoList = new ArrayList<ExtensionResourceInfo>();
        ExtensionResourceInfo extensionresourceinfo = new ExtensionResourceInfo();
        extensionresourceinfo.setModule("auth");
        extensionresourceinfo.setLocale("en");
        extensionresourceinfo.setData(keyAuthEn);
        extensionResourceInfoList.add(extensionresourceinfo);
        extensionresourceinfo = new ExtensionResourceInfo();
        extensionresourceinfo.setModule("auth");
        extensionresourceinfo.setLocale("zh_CN");
        extensionresourceinfo.setData(keyAuthZh);
        extensionResourceInfoList.add(extensionresourceinfo);
        return extensionResourceInfoList;
    }

    /**
     * get locale form vcetner
     * @param lang
     * @return
     */
    private String localeEdit(String lang) {
        String langStr = null;
        if (lang == null) {
            logger.info("Get the service instance languague which is null! ");
        }
        else if (lang.startsWith("zh")) {
            logger.info("Get the service instance languague which is zh! ");
            langStr = "zh";
        }
        else if (lang.startsWith("en")) {
            logger.info("Get the service instance languague which is en! ");
            langStr = "en";
        } else {
            langStr = lang;
            logger.info(String.format("Get the service instance languague which is {0}!", langStr));
        }
        return langStr;
    }
}
