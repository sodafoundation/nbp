// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package org.opensds.vmware.ngc.model;

public class DatastoreInfoOld {

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

    public boolean isCreateDatastore() {
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
