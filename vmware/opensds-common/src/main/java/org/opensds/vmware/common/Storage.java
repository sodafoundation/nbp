package org.opensds.vmware.common;

import org.opensds.vmware.common.models.*;

import java.util.List;

public abstract class Storage {
    protected String name;

    public Storage(String name) {
        this.name = name;
    }

    public abstract void login(String ip, int port, String user, String password) throws Exception;
    public abstract void logout();
    public abstract StorageMO getDeviceInfo() throws Exception;
    public abstract VolumeMO createVolume(String name, ALLOC_TYPE allocType, long capacity, String poolId) throws Exception;
    public abstract void deleteVolume(String volumeId) throws Exception;
    public abstract List<VolumeMO> listVolumes() throws Exception;
    public abstract List<VolumeMO> listVolumes(String poolId) throws Exception;
    public abstract List<StoragePoolMO> listStoragePools() throws Exception;
    public abstract StoragePoolMO getStoragePool(String poolId) throws Exception;
    public abstract void attachVolume(String volumeId, ConnectMO connect) throws Exception;
    public abstract void detachVolume(String volumeId, ConnectMO connect) throws Exception;
}
