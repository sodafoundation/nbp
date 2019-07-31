package com.opensds.esdk.utils;

import com.opensds.esdk.model.VCenterInfo;
import com.opensds.esdk.bean.ErrorCode;

import java.io.File;
import java.io.FileInputStream;


public class CommonUtils {

    private static final String HTTPS_HEAD = "https://";

    private static final String HTTP_HEAD = "http://";

    private static final String HTTP_SDK = "sdk";

    private static final String SEMICOLON = ":";

    private static final String SLANTING = "/";

    private static final String DIRECT_DOWNWLOAD_PATH = "download/";

    private static final String PACKAGE_NAME = "esdk.zip";

    private static final String PORT = "8088" ;

    /**
     * detect null
     * @param content
     * @return
     */
    public static boolean isNullStr (String content) {
        return (content == null|| content.isEmpty());
    }

    /**
     * check parmsters
     * @param vCenterInfo
     * @return
     */
    public static ErrorCode checkRegisterParameters(VCenterInfo vCenterInfo) {
        if(null == vCenterInfo || CommonUtils.isNullStr(vCenterInfo.getvCenterIp())
                || CommonUtils.isNullStr(vCenterInfo.getvCenterPassword())
                || CommonUtils.isNullStr(vCenterInfo.getvCenterUser()))
        {
            return ErrorCode.PARAM_IS_NULL;
        }
        return ErrorCode.SUCCESS;
    }

    /**
     * connect vcenter url
     * @param hostip
     * @return
     */
    public static String createVcUrl(String hostip) {
        return new StringBuffer().
                append(HTTPS_HEAD).
                append(hostip).
                append(SLANTING).
                append(HTTP_SDK).
                toString();
    }

    /**
     * create register download url
     * @param hostip
     * @return
     */
    public static String createRigesterUrl(String hostip) {
        return new StringBuffer().
                append(HTTPS_HEAD).
                append(hostip).
                append(SEMICOLON).
                append(PORT).
                append(SLANTING).
                append(DIRECT_DOWNWLOAD_PATH).
                append(PACKAGE_NAME).toString();
    }

    public static boolean isTheCorrectZipFile(String fileName) {
        return fileName == null ? false : fileName.equals(PACKAGE_NAME);
    }

    public static String getZipfilePath() {
        String zipPath = FileUtils.getProjectPath() + File.separator
                + "plugin" + File.separator + PACKAGE_NAME;
        return zipPath;
    }
}
