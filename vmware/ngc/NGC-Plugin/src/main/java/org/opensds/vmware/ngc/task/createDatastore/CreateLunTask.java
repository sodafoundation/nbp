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
import org.opensds.vmware.ngc.model.VolumeInfo;
import org.opensds.vmware.ngc.model.datastore.Datastore;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import org.opensds.vmware.ngc.models.ALLOC_TYPE;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.opensds.vmware.ngc.base.TaskInfoConst;
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
import java.util.Locale;
import java.util.Map;

public class CreateLunTask extends AbstractTask implements TaskExecution {

    private static Log logger = LogFactory.getLog(CreateLunTask.class);

    private VolumeMO volumeMO;

    private Map context;

    public CreateLunTask(Datastore datastore,
                         ManagedObjectReference[] hostMos,
                         ServerInfo serverInfo,
                         Storage storage) {
        super.serverInfo = serverInfo;
        super.hostMos = hostMos;
        super.datastore= datastore;
        super.storage = storage;
    }

    /**
     * set context
     * @param context Map
     */
    @Override
    public void setContext(Map context) {
        this.context = context;
    }

    /**
     * go run task to create volume
     * @throws Exception
     */
    @Override
    public void runTask() throws Exception {
        logger.info("---------Step one, run create the datastoreInfo task and create volume...");
        List<TaskInfo> taskInfoList = new ArrayList<>();
        try {
            createTaskList(taskInfoList, TaskInfoConst.Type.TASK_CREATE_LUN);
            VolumeInfo volumeInfo = ((VMFSDatastore)datastore).getVolumeInfos()[0];
            volumeMO = storage.createVolume(volumeInfo.getName(), volumeInfo.getDescription(),
                    volumeInfo.getAllocType().equals("thin") ? ALLOC_TYPE.THIN : ALLOC_TYPE.THICK,
                    CapacityUtil.convertCapToLong(volumeInfo.getCapacity()),
                    volumeInfo.getStoragePoolId());
            context.put(VolumeMO.class.getName(), volumeMO);
            changeTaskState(taskInfoList, TaskInfoConst.Status.SUCCESS,
                    String.format(Locale.ROOT, "Create volume %s finished.", volumeMO.name));
        } catch (Exception ex) {
            logger.error( "---------Step one error : " + ex.getMessage());
            changeTaskState(taskInfoList, TaskInfoConst.Status.ERROR,
                    String.format(Locale.ROOT, "Create volume failed: %s", ex.getMessage()));
            throw ex;
        }
    }

    @Override
    public void rollBack() throws Exception {
        if (volumeMO == null) {
            logger.info("Could not create the volume...");
            return;
        }
        logger.info("---------CreateLun/VolumeTask, rolling back and deleted created volume...");
        List<TaskInfo> rollBackTaskInfoList = new ArrayList<>();
        try {
            createTaskList(rollBackTaskInfoList, TaskInfoConst.Type.TASK_DELETE_LUN);
            storage.deleteVolume(volumeMO.id);
            changeTaskState(rollBackTaskInfoList, TaskInfoConst.Status.SUCCESS, "Delete volume finished.");
            for (ManagedObjectReference hostMo : hostMos) {
                hostServiceImpl.rescanAllHba(hostMo, serverInfo);
            }
        } catch (NullPointerException ex) {
            changeTaskState(rollBackTaskInfoList, TaskInfoConst.Status.ERROR,
                    String.format(Locale.ROOT, "Delete volume failed: %s", ex.getMessage()));
            throw ex;
        }
    }
}
