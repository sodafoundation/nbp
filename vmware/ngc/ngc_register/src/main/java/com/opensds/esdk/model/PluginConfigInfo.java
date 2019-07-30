package com.opensds.esdk.model;

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
