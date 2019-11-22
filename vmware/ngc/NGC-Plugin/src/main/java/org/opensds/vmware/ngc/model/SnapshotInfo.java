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

import org.opensds.vmware.ngc.models.SnapshotMO;
import org.opensds.vmware.ngc.util.CapacityUtil;

public class SnapshotInfo {

    private String id;

    private String name;

    private String status; // healthStatus in oceanstor

    private String storageId;

    private String parentId;

    private String capacity;

    private String timeStamp;  //active time in oceanstor

    public String getId() {
        return id;
    }

    public String getName() {
        return name;
    }

    public String getStatus() {
        return status;
    }

    public String getStorageId() {
        return storageId;
    }

    public String getParentId() {
        return parentId;
    }

    public void setId(String id) {
        this.id = id;
    }

    public void setName(String name) {
        this.name = name;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public void setDeviceId(String storageId) {
        this.storageId = storageId;
    }

    public void setParentId(String parentId) {
        this.parentId = parentId;
    }

    public String getCapacity() {
        return capacity;
    }

    public void setCapacity(String capacity) {
        this.capacity = capacity;
    }

    public String getTimeStamp() {
        return timeStamp;
    }

    public void setTimeStamp(String timeStamp) {
        this.timeStamp = timeStamp;
    }

    /**
     * convert snapshot Mo to snapshot info
     * @param snapshotMO
     * @return
     */
    public SnapshotInfo convertSnapShotMo2Info(SnapshotMO snapshotMO) {
        this.name = snapshotMO.name;
        this.id = snapshotMO.id;
        this.parentId = snapshotMO.parentID;
        this.status = snapshotMO.healthStatus;
        this.capacity = CapacityUtil.convert512BToCap(snapshotMO.capacity);
        this.timeStamp = snapshotMO.timeStamp;
        return this;
    }

    /**
     *  update the devie info
     * @param deviceInfo
     */
    public void updateWithStorage(DeviceInfo deviceInfo) {
        this.storageId = deviceInfo.uid;
    }


}
