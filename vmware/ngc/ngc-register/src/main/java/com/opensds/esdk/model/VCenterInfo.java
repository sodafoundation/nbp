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

package com.opensds.esdk.model;

import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.Size;

public class VCenterInfo {

    /**
     * ip of VSphere Web Client
     */
    @NotEmpty
    private String vCenterIp;

    /**
     * name of VSphere Web Client
     */
    @NotEmpty
    @Size(min=2, max=30)
    private String vCenterUser;

    /**
     * password of VSphere Web Client
     */
    @NotEmpty
    private String vCenterPassword;

    /**
     * the flag for register status
     * 1: already register ; 0: not regisiter
     */
    private String registerFlag;

    public String getvCenterIp() {
        return vCenterIp;
    }

    public void setvCenterIp(String vcIp) {
        this.vCenterIp = vcIp;
    }

    public String getvCenterUser() {
        return vCenterUser;
    }

    public void setvCenterUser(String vcUser) {
        this.vCenterUser = vcUser;
    }

    public String getvCenterPassword() {
        return vCenterPassword;
    }

    public void setvCenterPassword(String vcPassword) {
        this.vCenterPassword = vcPassword;
    }

    public String getRegisterFlag() {
        return registerFlag;
    }

    public void setRegisterFlag(String registerFlag) {
        this.registerFlag = registerFlag;
    }

    public VCenterInfo(String vCenterIp, String vCenterUser, String vCenterPassword) {
        this.vCenterIp = vCenterIp;
        this.vCenterUser = vCenterUser;
        this.vCenterPassword = vCenterPassword;
        this.registerFlag = "0";
    }

    public VCenterInfo(){
        this.vCenterIp = "";
        this.vCenterUser = "";
        this.vCenterPassword = "";
        this.registerFlag = "0";
    }

    @Override
    public String toString() {
        return "VcenterInfo [vCenterIp = " + vCenterIp + ", " +
                "vCenterUser = " + vCenterUser + ", " +
                "vCenterPassword = " + vCenterPassword + ", " +
                "registerFlag = " + registerFlag +
                "]";
    }
}
