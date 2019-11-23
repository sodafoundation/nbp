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

public class PluginConfigInfo {

    private String pluginName;

    private String pluginKey;

    private String pluginSummary;

    private String companyName;

    private String extensiontype;

    private String version;

    private String adminEmail;

    private String thumbprint;

    public String getThumbprint() {
        return thumbprint;
    }

    public void setThumbprint(String thumbprint) {
        this.thumbprint = thumbprint;
    }

    public String getPluginName() {
        return this.pluginName;
    }
    public void setPluginName(String pluginName) {
        this.pluginName = pluginName;
    }
    public String getPluginKey() {
        return this.pluginKey;
    }
    public void setPluginKey(String pluginKey) {
        this.pluginKey = pluginKey;
    }

    public String getPluginSummary() {
        return this.pluginSummary;
    }

    public void setPluginSummary(String summary) {
        this.pluginSummary = summary;
    }

    public String getCompanyName() {
        return this.companyName;
    }
    public void setCompanyName(String companyName) {
        this.companyName = companyName;
    }

    public String getExtensiontype() {
        return this.extensiontype;
    }
    public void setExtensiontype(String extensiontype) {
        this.extensiontype = extensiontype;
    }

    public String getVersion() {
        return this.version;
    }
    public void setVersion(String version) {
        this.version = version;
    }

    public String getAdminEmail() {
        return this.adminEmail;
    }
    public void setAdminEmail(String adminEmail) {
        this.adminEmail = adminEmail;
    }

    @Override
    public String toString() {
        return "PluginConfigInfo [pluginName =" + pluginName + ", " +
                "pluginSummary =" + pluginSummary + ", " +
                "pluginKey =" + pluginKey + ", " +
                "companyName =" + companyName +
                "extensiontype =" + companyName +
                "version =" + companyName +
                "]";
    }
}
