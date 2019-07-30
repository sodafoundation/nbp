package org.opensds.vmware.ngc.model;

public class HostInfo {
    private String connectedType;
    private String ip;
    private String name;

    public void setConnectedType(String isConnected) {
        this.connectedType = isConnected;
    }

    public String getConnectedType() {
        return connectedType;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }
    public String getIp() {
        return ip;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }
}
