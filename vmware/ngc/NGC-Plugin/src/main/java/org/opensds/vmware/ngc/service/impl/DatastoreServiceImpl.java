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

import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.base.DatastoreTypeEnum;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.service.DatastoreService;
import org.opensds.vmware.ngc.task.TaskExecution;
import org.opensds.vmware.ngc.task.TaskProcessor;
import org.opensds.vmware.ngc.task.createDatastore.CreateDatastoreTask;
import org.opensds.vmware.ngc.task.createDatastore.CreateLunTask;
import org.opensds.vmware.ngc.task.createDatastore.MountLunTask;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
public class DatastoreServiceImpl implements DatastoreService{

    private static final Log logger = LogFactory.getLog(DatastoreServiceImpl.class);

    @Autowired(required=false)
    private DeviceRepository deviceRepository;

    private static final String ERROR = "error";
    private static final String OK = "ok";

    @Override
    public ResultInfo<Object> createDatastore (ManagedObjectReference[] hostMos,
                                               ServerInfo serverInfo,
                                               DatastoreInfo datastoreInfo) {
        Storage storage = null;
        ResultInfo resultInfo = new ResultInfo();
        if (datastoreInfo.getDeviceId() != null) {
            DeviceInfo deviceInfo = deviceRepository.get(datastoreInfo.getDeviceId());
            storage = deviceRepository.getLoginedDeviceByIP(deviceInfo.ip);
        }
        if (storage == null) {
            resultInfo.setMsg("The device is not exist.");
            resultInfo.setStatus(ERROR);
            return resultInfo;
        }
        try {
            if (datastoreInfo.getDatastoreType().equals(DatastoreTypeEnum.VMFS_DATASTORE.getType())) {
                createVmfsDatastore(hostMos, serverInfo, datastoreInfo, storage);
            } else if (datastoreInfo.getDatastoreType().equals(DatastoreTypeEnum.NFS_DATASTORE.getType())) {
                createNfsDatastore(hostMos, serverInfo, datastoreInfo, storage);
            }
            resultInfo.setStatus(OK);
            return resultInfo;
        } catch (Exception e) {
            resultInfo.setMsg(e.getMessage());
            resultInfo.setStatus(ERROR);
            return resultInfo;
        }
    }


    /**
     * VMFS Datastore
     * @param hostMos
     * @param serverInfo
     * @param datastoreInfo
     * @param storage
     */
    public void createVmfsDatastore (ManagedObjectReference[] hostMos,
                                      ServerInfo serverInfo,
                                      DatastoreInfo datastoreInfo,
                                      Storage storage) {
        List<TaskExecution> taskList = new ArrayList<>();
        TaskExecution createLunTask = new CreateLunTask(datastoreInfo, hostMos, serverInfo, storage);
        taskList.add(createLunTask);
        TaskExecution mountLunTask = new MountLunTask(datastoreInfo, hostMos, serverInfo, storage);
        taskList.add(mountLunTask);
        if (datastoreInfo.isIsCreateDatastore()) {
            TaskExecution createDSTask = new CreateDatastoreTask(datastoreInfo, hostMos, serverInfo, storage);
            taskList.add(createDSTask);
        }
        TaskProcessor.runTaskWithThread(taskList);
    }

    /**
     * NFS Datastore
     * @param hostMos
     * @param serverInfo
     * @param datastoreInfo
     * @param device
     */
    private void createNfsDatastore (ManagedObjectReference[] hostMos,
                                     ServerInfo serverInfo,
                                     DatastoreInfo datastoreInfo,
                                     Storage storage) {
        logger.info("Not surpport nfs datastoreInfo!ï¼�ï¼�");
    }
}
