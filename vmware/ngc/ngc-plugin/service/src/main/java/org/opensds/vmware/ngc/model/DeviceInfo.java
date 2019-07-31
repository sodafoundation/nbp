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

import com.google.gson.Gson;
import java.io.Serializable;
import java.lang.reflect.Field;

public class DeviceInfo implements Serializable {

    public String ip;
    public String deviceName;
    public String deviceModel;
    public String deviceStatus;
    public String sn;
    public int port;
    public String deviceType;
    public String username;
    public String password;
    public String id;
    public String name;
    public String labelIds;
    public String primaryIconId;
    public String isRootFolder;
    private Object deviceReference;

    public DeviceInfo(String ip, String username, String password, int port, String deviceType, Object deviceReference) {
        this.ip = ip;
        this.username = username;
        this.password = password;
        this.port = port;
        this.deviceType = deviceType;
        this.deviceReference = deviceReference;
    }

    public DeviceInfo() {
    }

    public DeviceInfo(Object object) {
        deviceReference = object;
    }


    public Object getDeviceReference() {
        return deviceReference;
    }

    public void setDeviceReference(Object deviceReference) {
        this.deviceReference = deviceReference;
    }

    public String getProperty(String propName) {
        try {
            Field field = getClass().getDeclaredField(propName);
            field.setAccessible(true);
            return (String) field.get(this);
        } catch (Exception e) {
            //ignore it

        }
        return "";
    }

    @Override
    public String toString() {
        Gson json = new Gson();
        return json.toJson(this, DeviceInfo.class);
    }
}
