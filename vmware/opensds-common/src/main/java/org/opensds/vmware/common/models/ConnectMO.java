package org.opensds.vmware.common.models;

public class ConnectMO {
    public String name;
    public HOST_OS_TYPE osType;
    public String iscsiInitiator;
    public String[] fcInitiators;
    public ATTACH_MODE attachMode;
    public ATTACH_PROTOCOL attachProtocol;

    public ConnectMO(String name,
                     HOST_OS_TYPE osType,
                     String iscsiInitiator,
                     String[] fcInitiators,
                     ATTACH_MODE attachMode,
                     ATTACH_PROTOCOL attachProtocol) {
        this.name = name;
        this.osType = osType;
        this.iscsiInitiator = iscsiInitiator;
        this.fcInitiators = fcInitiators;
        this.attachMode = attachMode;
        this.attachProtocol = attachProtocol;
    }
}
