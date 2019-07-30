package org.opensds.vmware.ngc.base;


public enum CapabilityUnitTypeEnum {

    /**
     * byte
     */
    Byte(1L),

    /**
     * KB
     */
    KB(1024L),

    /**
     * MB
     */
    MB(1024L * 1024L),

    /**
     * GB
     */
    GB(1024L * 1024L * 1024L),

    /**
     * TB
     */
    TB(1024L * 1024L * 1024L * 1024L),

    /**
     * PB
     */
    PB(1024L * 1024L * 1024L * 1024L * 1024L);

    private long scale;

    CapabilityUnitTypeEnum(long scale)
    {
        this.scale = scale;
    }


    public long getUnit()
    {
        return this.scale;
    }


    public static CapabilityUnitTypeEnum getCapabilityUnitTypeByorder(int order)
    {
        for (CapabilityUnitTypeEnum one : CapabilityUnitTypeEnum.values())
        {
            if (one.ordinal() == order)
            {
                return one;
            }
        }
        return CapabilityUnitTypeEnum.MB;
    }
}
