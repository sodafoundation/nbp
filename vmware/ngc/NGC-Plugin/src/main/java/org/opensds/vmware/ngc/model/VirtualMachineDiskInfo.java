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

public class VirtualMachineDiskInfo implements Comparable<VirtualMachineDiskInfo>{

    private String name;

    private String id;

    private String size;

    private String diskFileName;

    private String diskMode;

    private Boolean isIndept;     // is rdm disk

    private String  datastoreName;

    private String datastoreId;

    private String lunIdentifier;

    private String lunUuid;

    private String compatibilityMode;

    public String getName() {
        return name;
    }

    public String getId() {
        return id;
    }

    public String getSize() {
        return size;
    }

    public String getDiskFileName() {
        return diskFileName;
    }

    public String getDiskMode() {
        return diskMode;
    }

    public Boolean getIndept() {
        return isIndept;
    }

    public String getDatastoreName() {
        return datastoreName;
    }

    public String getDatastoreId() {
        return datastoreId;
    }

    public String getLunIdentifier() {
        return lunIdentifier;
    }

    public String getLunUuid() {
        return lunUuid;
    }

    public void setName(String name) {
        this.name = name;
    }

    public void setId(String id) {
        this.id = id;
    }

    public void setSize(String size) {
        this.size = size;
    }

    public void setDiskFileName(String diskFileName) {
        this.diskFileName = diskFileName;
    }

    public void setDiskMode(String diskMode) {
        this.diskMode = diskMode;
    }

    public void setIndept(Boolean indept) {
        isIndept = indept;
    }

    public void setDatastoreName(String datastoreName) {
        this.datastoreName = datastoreName;
    }

    public void setDatastoreId(String datastoreId) {
        this.datastoreId = datastoreId;
    }

    public void setLunIdentifier(String lunIdentifier) {
        this.lunIdentifier = lunIdentifier;
    }

    public void setLunUuid(String lunUuid) {
        this.lunUuid = lunUuid;
    }
    public String getCompatibilityMode() {
        return compatibilityMode;
    }

    public void setCompatibilityMode(String compatibilityMode) {
        this.compatibilityMode = compatibilityMode;
    }

    @Override
    public int compareTo(VirtualMachineDiskInfo virDisk) {
        if (virDisk == null) {
            return 1;
        }
        if (this.getName().compareTo(virDisk.getName()) > 0) {
            return 1;
        }
        if (this.getName().compareTo(virDisk.getName()) < 0) {
            return -1;
        }
        if (this.equals(virDisk)) {
            return 0;
        }
        return 1;

    }
}
