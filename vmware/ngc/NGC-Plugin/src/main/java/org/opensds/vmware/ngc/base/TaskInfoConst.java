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

package org.opensds.vmware.ngc.base;

public interface TaskInfoConst {

    interface Type {

        String TASK_MOUNT_LUN_TO_HOST = "OpenSDS.Storage.Task.MountLunToHost";

        String TASK_UNMOUNT_LUN_FROM_HOST = "OpenSDS.Storage.Task.UnmountLunFromHost";

        String TASK_DELETE_LUN = "OpenSDS.Storage.Task.DeleteLun";

        String TASK_CREATE_LUN = "OpenSDS.Storage.Task.CreateLun";

        String TASK_CREATE_DATASTORE = "OpenSDS.Storage.Task.CreateDatastore";

        String TASK_CHECK_HOST_CONFIG = "OpenSDS.Storage.Task.CheckHostConfig";
    }

    interface Status {

        String ERROR = "ERROR";

        String SUCCESS = "SUCCESS";
    }
}
