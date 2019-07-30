package org.opensds.vmware.common.adapters.thirdparty.oceanstor;

enum ERROR_CODE {
    VOLUME_NOT_EXIST(1077936859);

    private long value;

    private ERROR_CODE(long v) {
        this.value = v;
    }

    public long getValue() {
        return this.value;
    }
}

enum RUNNING_STATUS {
    UNKNOWN(0), ONLINE(27), OFFLINE(28);

    private int value;

    private RUNNING_STATUS(int v) {
        this.value = v;
    }

    public int getValue() {
        return this.value;
    }
}