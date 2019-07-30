package org.opensds.vmware.ngc.model;



public class DatastoreInfo {
    /**
     * deviceType : OceanStorStorage
     * deviceId : urn:vri:Storage:210235980510F3000019
     * datastoreName : testDatastore1
     * storagePoolId : 0
     * datastoreType : lunDatastore
     * isCreateDatastore : true
     * fileVersion : VMFS5
     * lunName : testDatastore1
     * lunDescription :
     * lunCapacity : 3
     * allocType : thin
     */

    private String deviceType;
    private String deviceId;
    private String datastoreName;
    private String storagePoolId;
    private String datastoreType;
    private boolean isCreateDatastore;
    private String fileVersion;
    private String lunName;
    private String lunDescription;
    private long lunCapacity;
    private String allocType;

    public String getDeviceType() {
        return deviceType;
    }

    public void setDeviceType(String deviceType) {
        this.deviceType = deviceType;
    }

    public String getDeviceId() {
        return deviceId;
    }

    public void setDeviceId(String deviceId) {
        this.deviceId = deviceId;
    }

    public String getDatastoreName() {
        return datastoreName;
    }

    public void setDatastoreName(String datastoreName) {
        this.datastoreName = datastoreName;
    }

    public String getStoragePoolId() {
        return storagePoolId;
    }

    public void setStoragePoolId(String storagePoolId) {
        this.storagePoolId = storagePoolId;
    }

    public String getDatastoreType() {
        return datastoreType;
    }

    public void setDatastoreType(String datastoreType) {
        this.datastoreType = datastoreType;
    }

    public boolean isIsCreateDatastore() {
        return isCreateDatastore;
    }

    public void setIsCreateDatastore(boolean isCreateDatastore) {
        this.isCreateDatastore = isCreateDatastore;
    }

    public String getFileVersion() {
        return fileVersion;
    }

    public void setFileVersion(String fileVersion) {
        this.fileVersion = fileVersion;
    }

    public String getLunName() {
        return lunName;
    }

    public void setLunName(String lunName) {
        this.lunName = lunName;
    }

    public String getLunDescription() {
        return lunDescription;
    }

    public void setLunDescription(String lunDescription) {
        this.lunDescription = lunDescription;
    }

    public long getLunCapacity() {
        return lunCapacity;
    }

    public void setLunCapacity(int lunCapacity) {
        this.lunCapacity = lunCapacity;
    }

    public String getAllocType() {
        return allocType;
    }

    public void setAllocType(String allocType) {
        this.allocType = allocType;
    }
}
