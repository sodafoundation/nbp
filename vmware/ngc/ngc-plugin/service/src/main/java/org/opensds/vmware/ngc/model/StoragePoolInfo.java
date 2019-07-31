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

public class StoragePoolInfo {

    public static final String UNKNOW = "";

    public static final String ZERO = "0";

    private String id;

    private String name;

    private String parentType;

    private String parentId;

    private String parentName;

    private String healthStatus;

    private String runningStatus;

    private String description;

    private String workNodeId;

    private String totalCapacity;

    private String freeCapacity;

    private String availableCapacity;

    private String consumedCapacity;

    private String consumedCapacityPercentage;

    private String consumedCapacityThreshold;

    private String hotspareTotalCapacity;

    private String hotspareConsumedCapacity;

    private String hotspareConsumedCapacityPercentage;

    private String rawCapacity;

    private String replicationCapacity;

    private String sectorSize;

    private String usageType;

    public StoragePoolInfo() {
        super();
        this.id = UNKNOW;
        this.name = UNKNOW;
        this.parentType = UNKNOW;
        this.parentId = UNKNOW;
        this.parentName = UNKNOW;
        this.healthStatus = UNKNOW;
        this.runningStatus = UNKNOW;
        this.description = UNKNOW;
        this.workNodeId = UNKNOW;
        this.totalCapacity = ZERO;
        this.freeCapacity = ZERO;
        this.availableCapacity = ZERO;
        this.consumedCapacity = ZERO;
        this.consumedCapacityPercentage = ZERO;
        this.consumedCapacityThreshold = ZERO;
        this.hotspareTotalCapacity = ZERO;
        this.hotspareConsumedCapacity = ZERO;
        this.hotspareConsumedCapacityPercentage = ZERO;
        this.rawCapacity = ZERO;
        this.replicationCapacity = ZERO;
        this.sectorSize = ZERO;
    }


    public String PoolInfo() {
        return usageType;
    }

    public StoragePoolInfo setUsageType(String usageType) {
        this.usageType = usageType;
        return this;
    }

    public String getId() {
        return id;
    }

    public StoragePoolInfo setId(String id) {
        this.id = id;
        return this;
    }

    public String getName() {
        return name;
    }

    public StoragePoolInfo setName(String name) {
        this.name = name;
        return this;
    }

    public String getParentType() {
        return parentType;
    }

    public StoragePoolInfo setParentType(String parentType) {
        this.parentType = parentType;
        return this;
    }

    public String getParentId() {
        return parentId;
    }

    public StoragePoolInfo setParentId(String parentId) {
        this.parentId = parentId;
        return this;
    }

    public String getParentName() {
        return parentName;
    }

    public StoragePoolInfo setParentName(String parentName) {
        this.parentName = parentName;
        return this;
    }

    public String getHealthStatus() {
        return healthStatus;
    }

    public StoragePoolInfo setHealthStatus(String healthStatus) {
        this.healthStatus = healthStatus;
        return this;
    }

    public String getRunningStatus() {
        return runningStatus;
    }

    public StoragePoolInfo setRunningStatus(String runningStatus) {
        this.runningStatus = runningStatus;
        return this;
    }

    public String getDescription() {
        return description;
    }

    public StoragePoolInfo setDescription(String description) {
        this.description = description;
        return this;
    }

    public String getWorkNodeId() {
        return workNodeId;
    }

    public StoragePoolInfo setWorkNodeId(String workNodeId) {
        this.workNodeId = workNodeId;
        return this;
    }

    public String getTotalCapacity() {
        return totalCapacity;
    }

    public StoragePoolInfo setTotalCapacity(String totalCapacity) {
        this.totalCapacity = totalCapacity;
        return this;
    }

    public String getFreeCapacity() {
        return freeCapacity;
    }

    public StoragePoolInfo setFreeCapacity(String freeCapacity) {
        this.freeCapacity = freeCapacity;
        return this;
    }

    public String getAvailableCapacity() {
        return availableCapacity;
    }

    public StoragePoolInfo setAvailableCapacity(String availableCapacity) {
        this.availableCapacity = availableCapacity;
        return this;
    }

    public String getConsumedCapacity() {
        return consumedCapacity;
    }

    public StoragePoolInfo setConsumedCapacity(String consumedCapacity) {
        this.consumedCapacity = consumedCapacity;
        return this;
    }

    public String getConsumedCapacityPercentage() {
        return consumedCapacityPercentage;
    }

    public StoragePoolInfo setConsumedCapacityPercentage(String consumedCapacityPercentage) {
        this.consumedCapacityPercentage = consumedCapacityPercentage;
        return this;
    }

    public String getConsumedCapacityThreshold() {
        return consumedCapacityThreshold;
    }

    public StoragePoolInfo setConsumedCapacityThreshold(String consumedCapacityThreshold) {
        this.consumedCapacityThreshold = consumedCapacityThreshold;
        return this;
    }

    public String getHotspareTotalCapacity() {
        return hotspareTotalCapacity;
    }

    public StoragePoolInfo setHotspareTotalCapacity(String hotspareTotalCapacity) {
        this.hotspareTotalCapacity = hotspareTotalCapacity;
        return this;
    }

    public String getHotspareConsumedCapacity() {
        return hotspareConsumedCapacity;
    }

    public StoragePoolInfo setHotspareConsumedCapacity(String hotspareConsumedCapacity) {
        this.hotspareConsumedCapacity = hotspareConsumedCapacity;
        return this;
    }

    public String getHotspareConsumedCapacityPercentage() {
        return hotspareConsumedCapacityPercentage;
    }

    public StoragePoolInfo setHotspareConsumedCapacityPercentage(String hotspareConsumedCapacityPercentage) {
        this.hotspareConsumedCapacityPercentage = hotspareConsumedCapacityPercentage;
        return this;
    }

    public String getRawCapacity() {
        return rawCapacity;
    }

    public StoragePoolInfo setRawCapacity(String rawCapacity) {
        this.rawCapacity = rawCapacity;
        return this;
    }

    public String getReplicationCapacity() {
        return replicationCapacity;
    }

    public StoragePoolInfo setReplicationCapacity(String replicationCapacity) {
        this.replicationCapacity = replicationCapacity;
        return this;
    }

    public String getSectorSize() {
        return sectorSize;
    }

    public StoragePoolInfo setSectorSize(String sectorSize) {
        this.sectorSize = sectorSize;
        return this;
    }
}
