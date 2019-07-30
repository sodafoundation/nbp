package org.opensds.vmware.ngc.service;

import org.opensds.vmware.common.Storage;
import org.opensds.vmware.common.models.VolumeMO;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vim25.TaskInfo;
import com.vmware.vise.usersession.ServerInfo;



public interface HostService {

    TaskInfo createStorageTask(ManagedObjectReference hostMo, ServerInfo serverInfo, String taskId);

    Boolean changeTaskState(TaskInfo taskInfo, String taskState, String message);

    ResultInfo<Object> rescanAllHba(ManagedObjectReference host, ServerInfo serverInfo);

    ResultInfo<Object> mountVolume(ManagedObjectReference[] hostMos, ServerInfo serverInfo, Storage device, VolumeMO volumeMO);

    ResultInfo<Object> convertVmfsDatastore(ManagedObjectReference[] hostMos, ServerInfo serverInfo, VolumeMO volumeMO, DatastoreInfo datastoreInfo);

    ResultInfo<Object> getHostConnectionStateByHostMo(ManagedObjectReference hostMo, ServerInfo serverInfo);
}
