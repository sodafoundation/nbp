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

package org.opensds.vmware.ngc.controller;

import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.service.DataCenterService;
import org.opensds.vmware.ngc.service.DeviceService;
import org.opensds.vmware.ngc.service.HostService;
import org.opensds.vmware.ngc.util.QueryUtil;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.data.query.DataService;
import com.vmware.vise.data.query.ObjectReferenceService;
import com.vmware.vise.data.query.PropertyValue;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;


@Controller
@RequestMapping(value = "/data", method = RequestMethod.GET)
public class DataAccessController  {

    private static final Log logger = LogFactory.getLog(DataAccessController.class);

    private final static String OBJECT_ID = "id";
    @Autowired
    private DataService dataService;
    @Autowired
    private ObjectReferenceService objectReferenceService;
    @Autowired
    private HostService hostService;
    @Autowired
    private DataCenterService dataCenterService;
    @Autowired
    private DeviceService deviceService;

    public DataAccessController() {
        QueryUtil.setObjectReferenceService(objectReferenceService);
    }


    @RequestMapping(value = "/properties/{objectId}")
    @ResponseBody
    public Map<String, Object> getProperties(
            @PathVariable("objectId") String encodedObjectId,
            @RequestParam(value = "properties") String properties)
            throws Exception {

        Object ref = getDecodedReference(encodedObjectId);
        String objectId = objectReferenceService.getUid(ref);

        String[] props = properties.split(",");
        PropertyValue[] pvs = QueryUtil.getProperties(dataService, ref, props);
        Map<String, Object> propsMap = new HashMap<String, Object>();
        propsMap.put(OBJECT_ID, objectId);
        for (PropertyValue pv : pvs) {
            propsMap.put(pv.propertyName, pv.value);
        }
        return propsMap;
    }

    @RequestMapping(value = "/propertiesByRelation/{objectId}")
    @ResponseBody
    public PropertyValue[] getPropertiesForRelatedObject(
            @PathVariable("objectId") String encodedObjectId,
            @RequestParam(value = "relation") String relation,
            @RequestParam(value = "targetType") String targetType,
            @RequestParam(value = "properties") String properties)
            throws Exception {
        Object ref = getDecodedReference(encodedObjectId);
        String[] props = properties.split(",");
        PropertyValue[] result = QueryUtil.getPropertiesForRelatedObjects(
                dataService, ref, relation, targetType, props);
        return result;
    }

    private void ________________________________host_relation() {}

    @RequestMapping(value = "/hostList", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getHostListInfo(
            @RequestParam(value = "objectId") ManagedObjectReference ObjectMOR,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo){
        String type =  ObjectMOR.getType();
        if(type.equalsIgnoreCase(VimFieldsConst.MoTypesConst.Datacenter)){
            return dataCenterService.getHostListByDataCenterId(ObjectMOR ,serverInfo);
        }
        if(type.equalsIgnoreCase(VimFieldsConst.MoTypesConst.ClusterComputeResource)){
            return dataCenterService.getHostListByClusterId(ObjectMOR, serverInfo);
        }
        logger.error(" input is not datacenterName or clusterName!");
        return null;
    }


    @RequestMapping(value = "/hoststatus/{hostId}")
    @ResponseBody
    public ResultInfo<Object> getHostStatus(
            @PathVariable("hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo){
        ResultInfo<Object> resultInfo = hostService.getHostConnectionStateByHostMo(hostMoRef, serverInfo);
        return resultInfo;
    }

    private void ________________________________devices_relation() {}

    @RequestMapping(value = "/devicelist")
    @ResponseBody
    public ResultInfo<Object> getDeviceList()  {
        return deviceService.getList();
    }

    @RequestMapping(value = "/storagepoolListForBlock/{deviceId:.+}")
    @ResponseBody
    public ResultInfo<Object> getStoragePoolListForLun(
            @PathVariable("deviceId") String deviceId
    )  {
        return deviceService.getDeviceBlockPools(deviceId);
    }

    private void ________________________________others_relation() {}

    private Object getDecodedReference(String encodedObjectId) throws Exception {
        Object ref = objectReferenceService.getReference(encodedObjectId, true);
        if (ref == null) {
            throw new Exception("Object not found with id: " + encodedObjectId);
        }
        return ref;
    }
}

