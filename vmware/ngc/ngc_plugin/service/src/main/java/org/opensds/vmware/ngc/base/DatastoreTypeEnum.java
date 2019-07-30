package org.opensds.vmware.ngc.base;

public enum DatastoreTypeEnum {
    NFS_DATASTORE("nfsDatastore"),
    VMFS_DATASTORE("lunDatastore");

    private String type;

    public String getType() {
        return type;
    }

    DatastoreTypeEnum(String lunDatastore) {
        this.type = lunDatastore;
    }
}
