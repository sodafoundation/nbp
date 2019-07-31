package org.opensds.vmware.ngc.model.initiator;


public class StorageHostScsiInitiator extends StorageHostInitiator{

    private String wwpn;

    public String getWwpn() {
        return wwpn;
    }

    public void setWwpn(String wwpn) {
        this.wwpn = wwpn;
    }
}
