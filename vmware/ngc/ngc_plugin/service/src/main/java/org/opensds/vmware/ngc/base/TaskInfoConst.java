package org.opensds.vmware.ngc.base;



public interface TaskInfoConst {

    interface Type {

        String TASK_MOUNT_LUN_TO_HOST = "SODA.Storage.Task.MountLunToHost";

        String TASK_UNMOUNT_LUN_FROM_HOST = "SODA.Storage.Task.UnmountLunFromHost";

        String TASK_DELETE_LUN = "SODA.Storage.Task.DeleteLun";

        String TASK_CREATE_LUN = "SODA.Storage.Task.CreateLun";

        String TASK_CREATE_DATASTORE = "SODA.Storage.Task.CreateDatastore";

        String TASK_CHECK_HOST_CONFIG = "SODA.Storage.Task.CheckHostConfig";
    }

    interface Status {

        String ERROR = "ERROR";

        String SUCCESS = "SUCCESS";
    }
}
