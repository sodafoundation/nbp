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

package org.opensds.vmware.ngc.task.createDatastore;

import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.models.ALLOC_TYPE;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.opensds.vmware.ngc.base.TaskInfoConst;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import org.opensds.vmware.ngc.task.AbstractTask;
import org.opensds.vmware.ngc.task.TaskExecution;
import org.opensds.vmware.ngc.util.CapacityUtil;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vim25.TaskInfo;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;


import java.util.ArrayList;
import java.util.List;
import java.util.Map;


public class CreateLunTask extends AbstractTask implements TaskExecution {

    private static Log logger = LogFactory.getLog(CreateLunTask.class);

    private VolumeMO volumeMO;

    private Map context;

    public CreateLunTask(DatastoreInfo datastoreInfo,
                         ManagedObjectReference[] hostMos,
                         ServerInfo serverInfo,
                         Storage storage) {
        super.serverInfo = serverInfo;
        super.hostMos = hostMos;
        super.datastoreInfo = datastoreInfo;
        super.storage = storage;
    }

    @Override
    public void setContext(Map context) {
        this.context = context;
    }

    @Override
    public void runTask() throws Exception {

        logger.info("---------Step one, run create the datastoreInfo task and create volume...");
        List<TaskInfo> taskInfoList = new ArrayList<>();
        try {
            createTaskList(taskInfoList, TaskInfoConst.Type.TASK_CREATE_LUN);
            volumeMO = storage.createVolume(
                    datastoreInfo.getLunName(),
                    datastoreInfo.getAllocType().equals("thin")? ALLOC_TYPE.THIN :  ALLOC_TYPE.THICK,
                    CapacityUtil.converGBToByte(datastoreInfo.getLunCapacity()),
                    datastoreInfo.getStoragePoolId());
            context.put(VolumeMO.class.getName(), volumeMO);
            changeTaskState(taskInfoList,  TaskInfoConst.Status.SUCCESS,
                    String.format("Create volume %s finished.", volumeMO.name));
        } catch (Exception e) {
            logger.error( "---------Step one error : " + e.getMessage());
            changeTaskState(taskInfoList, TaskInfoConst.Status.ERROR,
                    String.format("Create volume failed: %s", e.getMessage()));
            throw e;
        }
    }

    @Override
    public void rollBack() throws Exception {
        if (volumeMO == null) {
            logger.info("Could not create the volume...");
            return;
        }
        logger.info("---------Step one, rolling back and deleted created volume...");
        List<TaskInfo> rollBackTaskInfoList = new ArrayList<>();
        try {
            createTaskList(rollBackTaskInfoList, TaskInfoConst.Type.TASK_DELETE_LUN);
            storage.deleteVolume(volumeMO.id);
            changeTaskState(rollBackTaskInfoList, TaskInfoConst.Status.SUCCESS, "Delete volume finished.");
            for (ManagedObjectReference hostMo : hostMos) {
                hostServiceImpl.rescanAllHba(hostMo, serverInfo);
            }
        } catch (Exception e) {
            changeTaskState(rollBackTaskInfoList, TaskInfoConst.Status.ERROR,
                    String.format("Delete volume failed: %s", e.getMessage()));
            throw e;
        }
    }

}
