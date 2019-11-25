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

import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.service.DeviceService;
import org.opensds.vmware.ngc.util.QueryUtil;
import com.vmware.vise.data.query.DataService;
import com.vmware.vise.data.query.ObjectReferenceService;
import com.vmware.vise.data.query.PropertyValue;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.ResponseBody;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestParam;
import java.util.HashMap;
import java.util.Map;

@Controller
@RequestMapping(value = "/data", method = RequestMethod.GET)
public class DataAccessController {

    private static final Log logger = LogFactory.getLog(DataAccessController.class);

    private static final String OBJECT_ID = "id";

    @Autowired
    private DataService dataService;
    @Autowired
    private ObjectReferenceService objectReferenceService;
    @Autowired
    private DeviceService deviceService;

    public DataAccessController() {
        QueryUtil.setObjectReferenceService(objectReferenceService);
    }

    /**
     * get properties from device
     * @param encodedObjectId String
     * @param properties String
     * @return Map<String, Object>
     * @throws Exception
     */
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
        Map<String, Object> propsMap = new HashMap<String, Object>(pvs.length + 1);
        propsMap.put(OBJECT_ID, objectId);
        for (PropertyValue pv : pvs) {
            propsMap.put(pv.propertyName, pv.value);
        }
        return propsMap;
    }

    /**
     * get properties related
     * @param encodedObjectId String
     * @param relation String
     * @param targetType String
     * @param properties String
     * @return PropertyValue
     * @throws Exception
     */
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


    /**
     * get storage pulllist
     * @param deviceId String
     * @return ResultInfo<Object>
     */
    @RequestMapping(value = "/storagepoolListForBlock/{deviceId:.+}")
    @ResponseBody
    public ResultInfo<Object> getStoragePoolListForLun(
            @PathVariable("deviceId") String deviceId) {
        return deviceService.getDeviceBlockPools(deviceId);
    }

    private Object getDecodedReference(String encodedObjectId) throws Exception {
        Object ref = objectReferenceService.getReference(encodedObjectId, true);
        if (ref == null) {
            throw new NullPointerException("Object not found with id: " + encodedObjectId);
        }
        return ref;
    }

}

