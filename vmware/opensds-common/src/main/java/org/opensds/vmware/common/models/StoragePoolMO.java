package org.opensds.vmware.common.models;

public class StoragePoolMO {
    public String name;
    public String id;
    public POOL_TYPE type;
    public long totalCapacity;
    public long freeCapacity;

    public StoragePoolMO(String name, String id, POOL_TYPE type, long totalCapacity, long freeCapacity) {
        this.name = name;
        this.id = id;
        this.type = type;
        this.totalCapacity = totalCapacity;
        this.freeCapacity = freeCapacity;
    }
}
