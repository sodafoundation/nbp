package org.opensds.vmware.common.models;

public class StorageMO {
    public String name;
    public String model;
    public String sn;
    public String status;
    public String vendor;

    public StorageMO(String name, String model, String sn, String status, String vendor) {
        this.name = name;
        this.model = model;
        this.sn = sn;
        this.status = status;
        this.vendor = vendor;
    }
}
