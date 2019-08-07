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

import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;
import org.opensds.vmware.ngc.base.TaskInfoConst;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import org.opensds.vmware.ngc.task.AbstractTask;
import org.opensds.vmware.ngc.task.TaskExecution;
import com.vmware.vim25.TaskInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.models.ConnectMO;
import org.opensds.vmware.ngc.models.VolumeMO;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class MountLunTask extends AbstractTask implements TaskExecution {

    private static Log logger = LogFactory.getLog(MountLunTask.class);

    private VolumeMO volumeMO;

    private ConnectMO connectMO;

    public MountLunTask(DatastoreInfo datastoreInfo,
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
        this.volumeMO = (VolumeMO)context.get(VolumeMO.class.getName());
    }

    @Override
    public void runTask() throws Exception {

        logger.info("---------Step two, run create the datastoreInfo task and mount volume...");

        List<TaskInfo> taskInfoList = new ArrayList<>();
        try {
            createTaskList(taskInfoList, TaskInfoConst.Type.TASK_MOUNT_LUN_TO_HOST);
            ResultInfo<Object> resultInfo = hostServiceImpl.mountVolume(hostMos, serverInfo, storage, volumeMO);
            if (resultInfo.getStatus().equals("ok")) {
                connectMO = (ConnectMO)resultInfo.getData();
                changeTaskState(taskInfoList, TaskInfoConst.Status.SUCCESS,
                        String.format("Mount volume %s fininsh.", datastoreInfo.getDatastoreName()));
            } else {
                changeTaskState(taskInfoList, TaskInfoConst.Status.ERROR,
                        String.format("Mount volume failed: %s", resultInfo.getMsg()));
                throw new Exception(resultInfo.getMsg());
            }
        } catch (Exception e) {
            changeTaskState(taskInfoList, TaskInfoConst.Status.ERROR,
                    String.format("Mount volume failed: %s", e.getMessage()));
            throw e;
        }
    }

    @Override
    public void rollBack() throws Exception {
        if (connectMO == null) {
            logger.info("---------Step two, do not need roll back for unmount volume.");
            return;
        }
        logger.info("---------Step two, rolling back mount volume...");
        List<TaskInfo> rollBackTaskInfoList = new ArrayList<>();
        try {
            createTaskList(rollBackTaskInfoList, TaskInfoConst.Type.TASK_UNMOUNT_LUN_FROM_HOST);
            storage.detachVolume(volumeMO.id, connectMO);
            changeTaskState(rollBackTaskInfoList, TaskInfoConst.Status.SUCCESS,
                    String.format("Umount volume %s finished.", volumeMO.name));
        } catch (Exception e) {
            changeTaskState(rollBackTaskInfoList, TaskInfoConst.Status.ERROR,
                    String.format("Unmount volume failed: %s", e.getMessage()));
            throw e;
        }
    }
}
