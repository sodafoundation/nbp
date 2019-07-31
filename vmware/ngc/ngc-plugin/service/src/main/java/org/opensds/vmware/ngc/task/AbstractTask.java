package org.opensds.vmware.ngc.task;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import org.opensds.vmware.ngc.service.HostService;
import org.opensds.vmware.ngc.service.impl.HostServiceImpl;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vim25.TaskInfo;
import com.vmware.vise.usersession.ServerInfo;
import org.opensds.vmware.common.Storage;
import java.util.List;


/**
 * Step 1 : @CreateLunTask
 * Step 2:  @MountLunTask
 * Step 3:  @CreateDatastoreTask
 */
public abstract class AbstractTask {

    protected HostService hostServiceImpl = HostServiceImpl.getInstance();

    private static Log logger = LogFactory.getLog(AbstractTask.class);

    protected ServerInfo serverInfo;

    protected ManagedObjectReference[] hostMos;

    protected DatastoreInfo datastoreInfo;

    protected Storage storage;


    protected AbstractTask() {
    }

    /**
     * create task in ESXI host
     * @param taskInfoList
     * @param taskType
     */
    protected void createTaskList(List<TaskInfo> taskInfoList, String taskType) {
        for (ManagedObjectReference hostMo :hostMos) {
            TaskInfo taskInfo = hostServiceImpl.createStorageTask(hostMo, serverInfo, taskType);
            taskInfoList.add(taskInfo);
        }
    }

    /**
     * change the task state
     * @param taskInfoList
     * @param taskStatus
     * @param msg
     */
    protected void changeTaskState(List<TaskInfo> taskInfoList, String taskStatus, String msg) {
        for (TaskInfo taskInfo : taskInfoList) {
            hostServiceImpl.changeTaskState(taskInfo, taskStatus, msg);
        }
    }
}
