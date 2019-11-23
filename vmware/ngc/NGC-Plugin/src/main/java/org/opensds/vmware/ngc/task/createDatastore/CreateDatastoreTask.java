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
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.model.VolumeInfo;
import org.opensds.vmware.ngc.model.datastore.Datastore;
import org.opensds.vmware.ngc.model.datastore.VMFSDatastore;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.opensds.vmware.ngc.base.TaskInfoConst;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.task.AbstractTask;
import org.opensds.vmware.ngc.task.TaskExecution;
import com.vmware.vim25.TaskInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;
import java.util.Map;

public class CreateDatastoreTask extends AbstractTask implements TaskExecution {

    private static Log logger = LogFactory.getLog(CreateDatastoreTask.class);

    private VolumeMO volumeMO;

    public CreateDatastoreTask(Datastore datastore,
                               ManagedObjectReference[] hostMos,
                               ServerInfo serverInfo,
                               Storage storage) {
        super.serverInfo = serverInfo;
        super.hostMos = hostMos;
        super.datastore = datastore;
        super.storage = storage;
    }

    /**
     * Set Context of volumeMo info
     * @param context Map
     */
    @Override
    public void setContext(Map context) {
        if (context.get(VolumeMO.class.getName()) instanceof VolumeMO) {
            this.volumeMO = (VolumeMO)context.get(VolumeMO.class.getName());
        }
    }

    /**
     * run task to create datastore
     * @throws Exception
     */
    @Override
    public void runTask() throws Exception {
        logger.info("---------CreateDatastoreTask, running create the datastoreInfo task and convert volume to Datastore...");
        List<TaskInfo> taskInfoList = new ArrayList<>();
        ResultInfo<Object> resultInfo = hostServiceImpl.convertVmfsDatastore(hostMos, serverInfo,
                volumeMO, (VMFSDatastore)datastore);
        if (resultInfo.getStatus().equals("ok")) {
            createTaskList(taskInfoList, TaskInfoConst.Type.TASK_CREATE_DATASTORE);
            changeTaskState(taskInfoList, TaskInfoConst.Status.SUCCESS,
                    String.format(Locale.ROOT, "Create Datastore %s finished.", volumeMO.name));
        } else if (resultInfo.getStatus().equals("error")) {
            changeTaskState(taskInfoList, TaskInfoConst.Status.ERROR,
                    String.format(Locale.ROOT, "Create Datastore %s failed.", volumeMO.name));
            throw new Exception(resultInfo.getMsg());
        } else {
            throw new IllegalArgumentException("Result status is illegal!");
        }
    }

    /**
     * fail to roll back
     * @throws Exception
     */
    @Override
    public void rollBack() throws Exception {
        logger.info("---------CreateDatastoreTask, roll back and create the datastore failed...");
    }
}
