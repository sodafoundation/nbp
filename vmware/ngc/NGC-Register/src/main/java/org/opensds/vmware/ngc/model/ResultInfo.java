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
