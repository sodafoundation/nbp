package org.opensds.vmware.ngc.model.initiator;


public class StorageHostIscsiInitiator extends StorageHostInitiator {

    private String iqn;

    public String getIqn()
    {
        return iqn;
    }

    public void setIqn(String iqn)
    {
        this.iqn = iqn;
    }
}
