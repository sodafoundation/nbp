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

public class ConnectMO {
    public String name;
    public HOST_OS_TYPE osType;
    public String iscsiInitiator;
    public String[] fcInitiators;
    public ATTACH_MODE attachMode;
    public ATTACH_PROTOCOL attachProtocol;

    public ConnectMO(String name,
                     HOST_OS_TYPE osType,
                     String iscsiInitiator,
                     String[] fcInitiators,
                     ATTACH_MODE attachMode,
                     ATTACH_PROTOCOL attachProtocol) {
        this.name = name;
        this.osType = osType;
        this.iscsiInitiator = iscsiInitiator;
        this.fcInitiators = fcInitiators;
        this.attachMode = attachMode;
        this.attachProtocol = attachProtocol;
    }
}
