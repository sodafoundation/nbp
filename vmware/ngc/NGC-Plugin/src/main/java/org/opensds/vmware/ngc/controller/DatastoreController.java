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
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.common.Request;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.datastore.Datastore;
import org.opensds.vmware.ngc.service.DatastoreService;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;
import org.opensds.vmware.ngc.service.Vmservice;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;

import static org.opensds.vmware.ngc.base.DatastoreTypeEnum.NFS_DATASTORE;
import static org.opensds.vmware.ngc.base.DatastoreTypeEnum.VMFS_DATASTORE;

@Controller
@RequestMapping(value = "/datastore")
public class DatastoreController {

    private static final Log logger = LogFactory.getLog(DatastoreController.class);

    @Autowired
    private DatastoreService datastoreService;

    /**
     * Create Datastore
     * @param hostMoRef ManagedObjectReference
     * @param serverInfo ServerInfo
     * @param json String
     * @return ResultInfo
     * @throws Exception
     */
    @RequestMapping(value = "/create", method = RequestMethod.POST)
    @ResponseBody
    public ResultInfo create(
            @RequestParam(value = "actionUid") String actionUid,
            @RequestParam(value = "objectId") ManagedObjectReference[] hostMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo,
            @RequestParam(value = "json") String json)
            throws Exception {
        Datastore datastore = convertToDatastore(json);
        ResultInfo resultInfo = datastoreService.create(hostMoRef, serverInfo, datastore);
        return resultInfo;
    }

    /**
     * Extend NFS/VMFS datastore size
     * @param serverInfo ServerInfo
     * @param json String
     * @return ResultInfo
     * @throws Exception
     */
    @RequestMapping(value = "/extend", method = RequestMethod.POST)
    @ResponseBody
    public ResultInfo extend(
            @RequestParam(value = "serverGuid") ServerInfo serverInfo,
            @RequestParam(value = "json") String json)
            throws Exception {
        Datastore datastore = convertToDatastore(json);
        ResultInfo resultInfo = datastoreService.extendSize(serverInfo, datastore);
        return resultInfo;
    }

    /**
     * Delete datastore
     * @param datastoreMo ManagedObjectReference
     * @param serverInfo ServerInfo
     * @return ResultInfo
     * @throws Exception
     */
    @RequestMapping(value = "/delete", method = RequestMethod.POST)
    @ResponseBody
    public ResultInfo delete(
            @RequestParam(value = "serverGuid") ServerInfo serverInfo,
            @RequestParam(value = "moref") ManagedObjectReference datastoreMo,
            @RequestParam(value = "json") String json)
            throws Exception {
        Datastore datastore = convertToDatastore(json);
        ResultInfo resultInfo = datastoreService.delete(serverInfo, datastoreMo, datastore);
        return resultInfo;
    }

    /**
     * Reduce nfs datastore size
     * @param serverInfo ServerInfo
     * @param json String
     * @return ResultInfo
     */
    @RequestMapping(value = "/nfsreduce", method = RequestMethod.POST)
    @ResponseBody
    public ResultInfo nfsreduce(
            @RequestParam(value = "serverGuid") ServerInfo serverInfo,
            @RequestParam(value = "json") String json) {
        Datastore datastoreInfo = convertToDatastore(json);
        // todo: Reduce the NFS datastore
        ResultInfo resultInfo = datastoreService.extendSize(serverInfo, datastoreInfo);
        return resultInfo;
    }

    /**
     * get volumes info with datastore
     * @param dsMo
     * @param serverInfo
     * @return
     */
    @RequestMapping(value = "/getInfo/{datastoreId}", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo getInfo(
            @PathVariable("datastoreId") ManagedObjectReference dsMo,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        ResultInfo resultInfo = datastoreService.getInfo(dsMo, serverInfo);
        return resultInfo;
    }


    // convert json to datastoreInfo
    public static Datastore convertToDatastore(String json) {
        logger.info("----------Begin convert mo to datastore!");
        Gson gson = new Gson();
        Datastore datastore = gson.fromJson(json, Datastore.class);
        if (datastore.getType().equals(VMFS_DATASTORE.getType())) {
            datastore = gson.fromJson(json, VMFSDatastore.class);
        } else if (datastore.getType().equals(NFS_DATASTORE.getType())) {
            datastore = gson.fromJson(json, VMFSDatastore.class);
        } else {
            throw new IllegalArgumentException("Datastore type is not surpport!");
        }
        return datastore;
    }
}
