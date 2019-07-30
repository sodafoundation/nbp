package org.opensds.vmware.ngc.task;


import java.util.Map;

public interface TaskExecution {
    void setContext(Map context);
    void runTask() throws Exception;
    void rollBack() throws Exception;
}
