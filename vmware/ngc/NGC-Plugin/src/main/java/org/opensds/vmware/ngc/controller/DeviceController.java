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

import com.google.gson.Gson;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.service.DeviceService;
import org.opensds.vmware.ngc.util.ObjectIdUtil;
import org.opensds.vmware.ngc.util.QueryUtil;
import com.vmware.vise.vim.data.VimObjectReferenceService;
import org.apache.commons.lang.StringUtils;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseBody;

@Controller
@RequestMapping(value = "/device")

public class DeviceController {

    private static final Log logger = LogFactory.getLog(DeviceController.class);

    @Autowired
    private DeviceService deviceService;
    @Autowired
    private VimObjectReferenceService objectReferenceService;

    public DeviceController() {
        QueryUtil.setObjectReferenceService(objectReferenceService);
    }

    /**
     * @param json String
     * @return ResultInfo
     */
    @RequestMapping(value = "/add", method = RequestMethod.POST)
    @ResponseBody
    public ResultInfo add(
            @RequestParam(value = "json") String json) {
        Gson gson = new Gson();
        DeviceInfo deviceInfo = gson.fromJson(json, DeviceInfo.class);
        ResultInfo resultInfo = deviceService.add(deviceInfo);
        return resultInfo;
    }

    /**
     * @param deviceID String
     * @return ResultInfo
     */
    @RequestMapping(value = "/delete", method = RequestMethod.DELETE)
    @ResponseBody
    public ResultInfo delete(
            @RequestParam(value = "deviceID") String deviceID) {
        Object objectReference = getObjectReference(deviceID);
        ResultInfo result = deviceService.delete(objectReference);
        return result;
    }

    /**
     * @param deviceID String
     * @return ResultInfo
     */
    @RequestMapping(value = "/get", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo get(
            @RequestParam(value = "deviceID") String deviceID) {
        Object objectReference = getObjectReference(deviceID);
        ResultInfo resultInfo = deviceService.get(objectReference);
        return resultInfo;
    }

    /**
     * @param deviceID String
     * @param json String
     * @return ResultInfo
     */
    @RequestMapping(value = "/update", method = RequestMethod.PUT)
    @ResponseBody
    public ResultInfo update(
            @RequestParam(value = "deviceID") String deviceID,
            @RequestParam(value = "json") String json) {
        Gson gson = new Gson();
        DeviceInfo deviceInfo = gson.fromJson(json, DeviceInfo.class);
        Object objectReference = getObjectReference(deviceID);
        ResultInfo resultInfo = deviceService.update(objectReference, deviceInfo);
        return resultInfo;
    }

    /**
     * get device list
     * @return ResultInfo<Object>
     */
    @RequestMapping(value = "/getList")
    @ResponseBody
    public ResultInfo<Object> getDeviceList() {
        return deviceService.getList();
    }

    /**
     * get device type
     * @return ResultInfo
     */
    @RequestMapping(value = "/types", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo types() {
        return deviceService.getAllDeviceType();
    }

    private Object getObjectReference(String targets) {
        if (StringUtils.isEmpty(targets)) {
            logger.error("Target is empty");
            return null;
        }
        String[] objectIDs = targets.split(",");
        if (objectIDs.length > 1) {
            logger.warn("The targets length exceed one ,will use the first one");
        }
        String objectID = ObjectIdUtil.decodeParameter(objectIDs[0]);
        Object objectReference = objectReferenceService.getReference(objectID);
        return objectReference;
    }
}

