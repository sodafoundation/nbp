package org.opensds.vmware.ngc.task;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;


public class TaskProcessor {
    private static Log logger = LogFactory.getLog(TaskProcessor.class);

    public static void runTaskWithThread(final List<TaskExecution> taskList) {
        Thread taskThread = new Thread(new Runnable() {
            @Override
            public void run() {
                final Map context= new HashMap();
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
        taskThread.run();
    }
}
