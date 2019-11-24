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

package org.opensds.vmware.ngc.models;

public class SnapshotMO {
    public String name;
    public String id;
    public String healthStatus;
    public long capacity;
    public String parentID;
    public String timeStamp;

    public SnapshotMO(String name, String id, String healthStatus, long capacity, String parentID, String timeStamp) {
        this.name = name;
        this.id = id;
        this.healthStatus = healthStatus;
        this.capacity = capacity;
        this.parentID = parentID;
        this.timeStamp = timeStamp;
    }
}
