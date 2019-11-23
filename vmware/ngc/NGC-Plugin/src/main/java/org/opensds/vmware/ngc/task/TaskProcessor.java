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

package org.opensds.vmware.ngc.task;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class TaskProcessor {

    private static final Log logger = LogFactory.getLog(TaskProcessor.class);

    /**
     * Run Task List
     * @param taskList List<TaskExecution>
     */
    public static void runTaskWithThread(final List<TaskExecution> taskList) {
        Thread taskThread = new Thread(new Runnable() {
            @Override
            public void run() {
                final Map context = new HashMap();
                List<TaskExecution> taskStack = new ArrayList<>();
                try {
                    try {
                        for (TaskExecution task : taskList) {
                            taskStack.add(task);
                            task.setContext(context);
                            task.runTask();
                        }
                    } catch (Exception e) {
                        logger.error(e.getMessage());
                        for (int i = taskStack.size() - 1; i >= 0; i--) {
                            TaskExecution reversTask = taskStack.get(i);
                            reversTask.rollBack();
                        }
                    }
                } catch (Exception ex) {
                    logger.error(ex.getMessage());
                }
            }
        });
        taskThread.start();
    }
}
