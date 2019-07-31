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
