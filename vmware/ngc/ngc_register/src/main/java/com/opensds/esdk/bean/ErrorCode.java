package com.opensds.esdk.bean;

import org.springframework.stereotype.Component;


public final class ErrorCode {

    /**
     * errorCode msg;
     */
    private Long errodCode;
    /**
     * errorDESC msg;
     */
    private String errorDESC;

    private ErrorCode(Long errodCode, String errorDESC) {
        this.errodCode = errodCode;
        this.errorDESC = errorDESC;
    }

    /**
     * register success
     */
    public static final ErrorCode SUCCESS = new ErrorCode(0L, "Success!");

    /**
     * connect fail
     */
    public static final ErrorCode CONNECT_FAIL = new ErrorCode(1L, "Connect fail!");

    /**
     * unsupport localses
     */
    public static final ErrorCode INVALID_LOGIN = new ErrorCode(2L, "Password or UserName is invalid!");
    /**
     * supported locales
     */
    public static final ErrorCode UNSUPPORT_LOCALSES = new ErrorCode(3L, "Do not Support this language!");

    /**
     * already register in plugin
     */
    public static final ErrorCode ALREADY_REGISTER = new ErrorCode(5L, "Already register the plugin!");

    /**
     * not register in plugin
     */
    public static final ErrorCode NOT_ALREADY_REGISTER = new ErrorCode(6L, "The plugin is not registered!");

    /**
     * Space .     */
    public static final ErrorCode NO_SPACE = new ErrorCode(7L, "No space in the disk!");

    /**
     *  Write config failed
     */
    public static final ErrorCode WRITE_CONFIG_FAILED = new ErrorCode(8L, "Write config file failed!");

    /**
     * Plugin properties has wrong
     */
    public static final ErrorCode PLUGIN_PROPERTIES_FAILED = new ErrorCode(9L, "Plugin properties has some thing wrong!");

    /**
     * Can not get vcenter session ip
     */
    public static final ErrorCode CONNECT_FAIL_GET_LOCAL_IP = new ErrorCode(10L, "Can not get the local machine session IP!");

    /**
     * parameter is null
     */
    public static final ErrorCode PARAM_IS_NULL = new ErrorCode(11100001L, "nput parameter is null!");

    public Long getErrodCode() {
        return errodCode;
    }

    public String getErrorDESC() {
        return errorDESC;
    }
}
