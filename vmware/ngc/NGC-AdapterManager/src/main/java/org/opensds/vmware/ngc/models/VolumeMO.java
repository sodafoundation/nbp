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

public class VolumeMO {

	public String name;
    public String id;
    public String wwn;
    public ALLOC_TYPE allocType;
    public long capacity;

    //add 0903
    public long allocCapacity;
    public StatusE status;
    public String storagePoolId;

    public VolumeMO(String name, String id, String wwn, ALLOC_TYPE allocType, long capacity) {
        this.name = name;
        this.id = id;
        this.wwn = wwn;
        this.allocType = allocType;
        this.capacity = capacity;
    }

    @Override
	public String toString() {
		return "VolumeMO [name=" + name + ", id=" + id + ", wwn=" + wwn + ", allocType=" + allocType + ", capacity="
				+ capacity + ", allocCapacity=" + allocCapacity + ", status=" + status + ", storagePoolId="
				+ storagePoolId + "]";
	}

    public static enum StatusE {

        Normal(1),
        Faulty(2);

        private int value;
        StatusE(int  value) {
            this.value = value;
        }
    }
}

