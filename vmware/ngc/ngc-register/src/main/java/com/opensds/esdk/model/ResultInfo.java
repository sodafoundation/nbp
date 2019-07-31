package com.opensds.esdk.model;

import org.apache.http.HttpStatus;

public class ResultInfo<T> {

    private Integer errorCode;

    private String errorDescription;

    private T data;

    public ResultInfo(T data) {
        this.data = data;
    }

    public ResultInfo() {
    }
    public ResultInfo(Integer erroCode, String errorDescription) {
        this.errorCode = erroCode;
        this.errorDescription = errorDescription;
    }

    public Integer getErrorCode() {
        return errorCode;
    }

    public ResultInfo setErrorCode(Integer errorCode) {
        this.errorCode = errorCode;
        return this;
    }

    public String getErrorDesc() {
        return errorDescription;
    }

    public ResultInfo setErrorDesc(String errorDesc) {
        if (this.errorCode == null) {
            this.errorCode = HttpStatus.SC_INTERNAL_SERVER_ERROR;
        }
        this.errorDescription = errorDesc;
        return this;
    }

    public T getData() {
        return data;
    }

    public ResultInfo setData(T data) {
        this.data = data;
        return this;
    }

    public String getErrorCodemsg() {
        return "Result error [errorCode: " + getErrorCode().toString()
                + "; errorDescription : " + getErrorDesc()
                + "]";
    }
}
