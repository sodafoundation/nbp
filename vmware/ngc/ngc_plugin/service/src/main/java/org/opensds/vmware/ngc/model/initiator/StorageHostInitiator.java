package org.opensds.vmware.ngc.model.initiator;

import org.opensds.vmware.ngc.base.HostHbaEnum;


public abstract class StorageHostInitiator {

    private HostHbaEnum hbaType;

    public HostHbaEnum getHbaType()
    {
        return hbaType;
    }

    public void setHbaType(HostHbaEnum hbaType)
    {
        this.hbaType = hbaType;
    }
}
