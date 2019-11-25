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

package org.opensds.vmware.ngc.model.datastore;

public class Datastore {

    private String name;

    private String ID;

    private Boolean accessible;

    // vmfs  || nfs || vvol
    private String type;

    private String capacity;

    private String freeCapacity;

    private double capUsage;

    private Long extendCapacciy;

    private boolean isCreateDatastore;

    private String overallStatus;

    private String hardWareAccSupport;

    private String lastTime;

    public void setName(String datastoreName) {
        this.name = datastoreName;
    }

    public String getName() {
        return name;
    }

    public String getType() {
        return type;
    }

    public void setType(String datastoreType) {
        this.type = datastoreType;
    }

    public String getCapacity() {
        return capacity;
    }

    public void setCapacity(String datastoreCapacity) {
        this.capacity = datastoreCapacity;
    }

    public void setId(String ID) {
        this.ID = ID;
    }

    public String getId() {
        return ID;
    }

    public boolean isCreateDatastore() {
        return isCreateDatastore;
    }

    public void setCreateDatastore(boolean createDatastore) {
        isCreateDatastore = createDatastore;
    }

    public Long getExtendCapacciy() {
        return extendCapacciy;
    }

    public void setExtendCapacciy(Long extendCapacciy) {
        this.extendCapacciy = extendCapacciy;
    }

    public String getFreeCapacity() {
        return freeCapacity;
    }

    public void setFreeCapacity(String freeCapacity) {
        this.freeCapacity = freeCapacity;
    }

    public String getOverallStatus() {
        return overallStatus;
    }

    public void setOverallStatus(String overallStatus) {
        this.overallStatus = overallStatus;
    }

    public String getHardWareAccSupport() {
        return hardWareAccSupport;
    }

    public void setHardWareAccSupport(String hardWareAccSurport) {
        this.hardWareAccSupport = hardWareAccSurport;
    }

    public Boolean getAccessible() {
        return accessible;
    }

    public void setAccessible(Boolean accessible) {
        this.accessible = accessible;
    }

    public String getLastTime() {
        return lastTime;
    }

    public void setLastTime(String lastTime) {
        this.lastTime = lastTime;
    }


    public double getCapUsage() {
        return capUsage;
    }

    public void setCapUsage(double capUsage) {
        this.capUsage = capUsage;
    }


}
