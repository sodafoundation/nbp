package org.opensds.vmware.ngc.base;

public enum HostHbaEnum {
    ISCSI("ISCSI"),
    FC("FC"),
    FCOE("FCOE");

    private String hbaType;

    public String getHbaType() {
        return hbaType;
    }

    HostHbaEnum(String type) {
        this.hbaType = type;
    }
}
