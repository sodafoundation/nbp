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

package org.opensds.vmware.ngc.utils;

import org.opensds.vmware.ngc.model.AuthInfo;
import org.opensds.vmware.ngc.model.VCenterInfo;
import org.opensds.vmware.ngc.bean.ErrorCode;
import org.opensds.vmware.ngc.model.EventsInfo;
import org.opensds.vmware.ngc.model.TaskInfo;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.dom4j.Document;
import org.dom4j.DocumentHelper;
import org.dom4j.Element;
import org.dom4j.Node;
import org.dom4j.io.SAXReader;
import org.dom4j.io.XMLWriter;

import java.io.*;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

public class XmlUtils {

    private static final Logger _logger = LogManager.getLogger(XmlUtils.class);

    public static XmlUtils xmlUitlObj;

    public static final String BASE_PATH = "static/config/";

    public static final String EVENTTYPEINFO = BASE_PATH +  "EventTypeInfo";

    public static final String EVENTTYPEINFO_EN = BASE_PATH  + "EventTypeInfo_en.xml";

    public static final String EVENTTYPEINFO_ZH = BASE_PATH  + "EventTypeInfo_zh.xml";

    public static final String TASKINFO = BASE_PATH  + "TaskInfo";

    public static final String TASKINFO_EN = BASE_PATH + "TaskInfo_en.xml";

    public static final String TASKINFO_ZH = BASE_PATH + "TaskInfo_zh.xml";

    public static final String AUTHINFO = BASE_PATH + "AuthInfo";

    public static final String AUTHINFO_EN = BASE_PATH + "AuthInfo_en.xml";

    public static final String AUTHINFO_ZH = BASE_PATH + "AuthInfo_zh.xml";

    public static final String VCENTER_CONFIG = FileUtils.getProjectPath() +
            File.separator + "conf" + File.separator+ "PluginVCenterConfig.xml";


    public static synchronized XmlUtils getXml() {
        if (null == xmlUitlObj) {
            xmlUitlObj = new XmlUtils();
        }
        return xmlUitlObj;
    }

    public List<AuthInfo> xmlReadAuthInfo(String local) {
        String authInfoXml = AUTHINFO;
        if (null == local) {
            _logger.info("the local is null when read authInfoXml");
            return null;
        }
        authInfoXml = AUTHINFO + '_' + local + ".xml";
        Document doc = load(authInfoXml);
        return xmlReadAuthInfo(doc, local);
    }

    public List<AuthInfo> xmlReadAllAuthInfo() {
        Document docen = load(AUTHINFO_EN);
        Document doczh = load(AUTHINFO_ZH);
        List<AuthInfo> authInfoList = xmlReadAuthInfo(docen, "en");
        List<AuthInfo> authInfoListZh = xmlReadAuthInfo(doczh, "zh");
        authInfoList.addAll(authInfoListZh);
        return authInfoList;
    }

    public List<AuthInfo> xmlReadAuthInfo(Document document, String local) {
        Document doc = document;
        List<Node> list = doc.selectNodes("/ROOT/AuthId");
        Iterator<Node> it = list.iterator();
        List<AuthInfo> authInfoList = new ArrayList<AuthInfo>();
        while (it.hasNext()) {
            AuthInfo authInfo = new AuthInfo();
            Element taskElement = (Element) it.next();
            Node authName = taskElement.selectSingleNode("authName");
            Node authNaLabelKey = taskElement.selectSingleNode("authNaLabelKey");
            Node authNaLabelValue = taskElement.selectSingleNode("authNaLabelValue");
            Node authNaSummaryKey = taskElement.selectSingleNode("authNaSummaryKey");
            Node authNaSummaryValue = taskElement.selectSingleNode("authNaSummaryValue");
            Node authId = taskElement.selectSingleNode("authId");
            Node authIdLabelKey = taskElement.selectSingleNode("authIdLabelKey");
            Node authIdLabelValue = taskElement.selectSingleNode("authIdLabelValue");
            Node authIdSummaryKey = taskElement.selectSingleNode("authIdSummaryKey");
            Node authIdSummaryValue = taskElement.selectSingleNode("authIdSummaryValue");
            authInfo.setAuthName(authName.getText());
            authInfo.setAuthNaLabel(authNaLabelKey.getText());
            authInfo.setAuthNaLabelValue(authNaLabelValue.getText());
            authInfo.setAuthNaSummary(authNaSummaryKey.getText());
            authInfo.setAuthNaSummaryValue(authNaSummaryValue.getText());
            authInfo.setAuthId(authId.getText());
            authInfo.setAuthIdLabel(authIdLabelKey.getText());
            authInfo.setAuthIdLabelValue(authIdLabelValue.getText());
            authInfo.setAuthIdSummary(authIdSummaryKey.getText());
            authInfo.setAuthIdSummaryValue(authIdSummaryValue.getText());
            authInfo.setAuthLocal(local);
            authInfoList.add(authInfo);
        }
        return authInfoList;
    }

    public List<TaskInfo> xmlReadTaskInfo(String local) {
        String taskInfoXml = TASKINFO;
        if (null == local) {
            _logger.info("the local is null when read taskInfoXml");
            return null;
        }
        taskInfoXml = TASKINFO + '_' + local + ".xml";
        Document doc = load(taskInfoXml);
        return xmlReadTaskInfo(doc, local);
    }

    public List<TaskInfo> xmlReadAllTaskInfo() {
        Document docen = load(TASKINFO_EN);
        Document doczh = load(TASKINFO_ZH);
        List<TaskInfo> taskInfoList = xmlReadTaskInfo(docen, "en");
        List<TaskInfo> taskInfoListZh = xmlReadTaskInfo(doczh, "zh");
        taskInfoList.addAll(taskInfoListZh);
        return taskInfoList;
    }

    public List<TaskInfo> xmlReadTaskInfo(Document document, String local) {
        Document doc = document;
        List<Node> list = doc.selectNodes("/ROOT/TaskId");
        Iterator<Node> it = list.iterator();
        List<TaskInfo> taskInfoList = new ArrayList<TaskInfo>();
        while (it.hasNext()) {
            TaskInfo taskInfo = new TaskInfo();
            Element taskElement = (Element) it.next();
            Node taskId = taskElement.selectSingleNode("taskId");
            Node labelKey = taskElement.selectSingleNode("taskLabelKey");
            Node labelValue = taskElement.selectSingleNode("taskLabelValue");
            Node summaryKey = taskElement.selectSingleNode("taskSummaryKey");
            Node summaryValue = taskElement.selectSingleNode("taskSummaryValue");
            taskInfo.setTaskId(taskId.getText());
            taskInfo.setTaskLabel(labelKey.getText());
            taskInfo.setTaskLabelValue(labelValue.getText());
            taskInfo.setTaskSummary(summaryKey.getText());
            taskInfo.setTaskSummaryValue(summaryValue.getText());
            taskInfo.setTasklocal(local);
            taskInfoList.add(taskInfo);
        }
        return taskInfoList;
    }

    public List<EventsInfo> xmlReadEventTypeInfo(String local) {
        String eventTypeInfoXml = EVENTTYPEINFO;
        if (null == local) {
            _logger.error("The eventtypeinfo is null");
            return null;
        }
        eventTypeInfoXml = EVENTTYPEINFO + '_' + local + ".xml";
        Document doc = load(eventTypeInfoXml);
        return xmlReadEventTypeInfo(doc, local);
    }

    public List<EventsInfo> xmlReadAllEventInfo() {
        Document docen = load(EVENTTYPEINFO_EN);
        Document doczh = load(EVENTTYPEINFO_ZH);
        List<EventsInfo> eventInfoList = xmlReadEventTypeInfo(docen,
                "en");
        List<EventsInfo> eventInfoListZh = xmlReadEventTypeInfo(doczh,
                "zh");
        eventInfoList.addAll(eventInfoListZh);
        return eventInfoList;
    }

    public List<EventsInfo> xmlReadEventTypeInfo(Document document, String local) {
        Document doc = document;
        List<Node> list = doc.selectNodes("/ROOT/EventID");
        Iterator<Node> it = list.iterator();
        List<EventsInfo> eventInfoList = new ArrayList<EventsInfo>();
        while (it.hasNext()) {
            EventsInfo event = new EventsInfo();
            Element eventElement = (Element) it.next();
            Node evety = eventElement.selectSingleNode("EventType");
            Node id = eventElement.selectSingleNode("EventType/eventTypeID");
            Node des = eventElement.selectSingleNode("EventType/description");
            event.setEventName(String.valueOf(eventElement.attribute("name")
                    .getValue()));
            event.setEventTypeId(id.getText());
            event.setEventDescription(des.getText());
            event.setEventTypeSchema(evety.asXML());
            event.setEventSeverity(eventElement.attribute("severity")
                    .getValue());
            event.setEventLocal(local);
            eventInfoList.add(event);
        }
        return eventInfoList;
    }

    public VCenterInfo xmlReadVcenterInfo() {
        return xmlReadVcenterInfo(VCENTER_CONFIG);
    }

    public VCenterInfo xmlReadVcenterInfo(String path) {
        VCenterInfo vcInfo = new VCenterInfo();
        try {
            if (path == null) {
                path = VCENTER_CONFIG;
            }
            Document doc = loadConigXml(path);
            Element root = doc.getRootElement();
            Iterator< ? > it = root.elementIterator();

            while (it.hasNext()) {
                Element vcElement = (Element) it.next();
                String vcip = vcElement.attributeValue("vcip");
                String vcname = vcElement.attributeValue("vcname");
                String vcpwd = vcElement.attributeValue("vcpwd");
                String flag = vcElement.attributeValue("registerflag");
                vcInfo.setvCenterIp(vcip);
                vcInfo.setvCenterUser(vcname);
                vcInfo.setvCenterPassword(vcpwd);
                vcInfo.setRegisterFlag(flag);
            }

        } catch (Exception ex) {
            _logger.error("Read XML error, ex : " + ex.getMessage());
        }
        return vcInfo;
    }

    public ErrorCode xmlWriteVcenterInfo(VCenterInfo vc) {
        Document doc = loadConigXml(VCENTER_CONFIG);
        Element root = doc.getRootElement();
        Iterator< ? > it = root.elementIterator();
        while (it.hasNext())
        {
            Element vcElement = (Element) it.next();
            vcElement.attribute("vcip").setText(vc.getvCenterIp());
            vcElement.attribute("vcname").setText(vc.getvCenterUser());
            vcElement.attribute("vcpwd").setText(vc.getvCenterPassword());
            vcElement.attribute("registerflag").setText(vc.getRegisterFlag());
        }
        return writeToXML(VCENTER_CONFIG, doc);
    }

    private ErrorCode writeToXML(String path, Document doc) {
        if (OperationSystemUtils.isFreeDiskSpace()) {
            return ErrorCode.NO_SPACE;
        }
        OutputStream outputStream = null;
        Writer writer = null;
        XMLWriter xmlWriter = null;
        try {
            outputStream = new FileOutputStream(path);
            writer = new OutputStreamWriter(outputStream, "UTF-8");
            xmlWriter = new XMLWriter(writer);
            xmlWriter.write(doc);
            xmlWriter.flush();
            return ErrorCode.SUCCESS;
        } catch (IOException e) {
            _logger.error("Writer xml error ;" + e.toString());
            return ErrorCode.WRITE_CONFIG_FAILED;
        } finally {
            if (xmlWriter != null) {
                try {
                    xmlWriter.close();
                }
                catch (IOException e) {
                    _logger.error("XMLWriter close error" + e.toString());
                }
            }
            closeStream(writer);
            closeStream(outputStream);
        }
    }

    public Document loadConigXml(String filename) {
        Document document = null;
        Reader xmlReader = null;
        FileInputStream stream = null;
        try {
            SAXReader saxReader = new SAXReader();
            stream =  new FileInputStream(new File(filename));
            xmlReader = new InputStreamReader(stream, "UTF-8");
            document = saxReader.read(xmlReader);
        }
        catch (Exception ex) {
            _logger.error(String.format("the filename of document is %s, error [%s]" , filename
                    ,ex.getMessage()));
            document = DocumentHelper.createDocument();
            document.addElement("ROOT");
        } finally {
            closeStream(xmlReader);
            closeStream(stream);
        }
        return document;
    }


    public Document load(String filename) {
        Document document = null;
        Reader xmlReader = null;
        InputStream stream = null;
        try {
            SAXReader saxReader = new SAXReader();
            stream = getClass().getClassLoader().getResourceAsStream(filename);
            xmlReader = new InputStreamReader(stream, "UTF-8");
            document = saxReader.read(xmlReader);
        }
        catch (Exception ex) {
            _logger.error(String.format("the filename of document is %s, error [%s]" , filename
                    ,ex.getMessage()));
            document = DocumentHelper.createDocument();
            document.addElement("ROOT");
        } finally {
            closeStream(xmlReader);
            closeStream(stream);
        }
        return document;
    }

    private void closeStream(Closeable stream) {
        if (stream != null) {
            try {
                stream.close();
            }
            catch (IOException e) {
                _logger.error(e.getMessage(), e);
            }
        }
    }
}
