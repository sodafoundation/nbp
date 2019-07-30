package org.opensds.vmware.common.models;

public class VolumeMO {
    public String name;
    public String id;
    public String wwn;
    public ALLOC_TYPE allocType;
    public long capacity;

    public VolumeMO(String name, String id, String wwn, ALLOC_TYPE allocType, long capacity) {
        this.name = name;
        this.id = id;
        this.wwn = wwn;
        this.allocType = allocType;
        this.capacity = capacity;
    }
}
