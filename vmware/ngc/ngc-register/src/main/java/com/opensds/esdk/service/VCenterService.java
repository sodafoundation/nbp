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

package com.opensds.esdk.service;

import com.opensds.esdk.model.ResultInfo;
import com.opensds.esdk.model.VCenterInfo;
import com.opensds.esdk.bean.ErrorCode;
import com.opensds.esdk.utils.CommonUtils;
import com.opensds.esdk.utils.XmlUtils;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class VCenterService {

    private static final Logger logger = LogManager.getLogger(VCenterService.class);

    @Autowired
    private PluginMgntService pluginMgntService;

    /**
     * connect with vcenter server and register the plugin
     * @param vCenterInfo
     * @return
     */
    public ResultInfo registerPlugin(VCenterInfo vCenterInfo) {
        logger.info("-----------------Begin regiser the plugin!-----------------");
        ResultInfo resultInfo = new ResultInfo();
        ErrorCode chResult = pluginMgntService.register(vCenterInfo);
        resultInfo.setErrorCode(chResult.getErrodCode().intValue());
        resultInfo.setErrorDesc(chResult.getErrorDESC());
        if (ErrorCode.SUCCESS.getErrodCode() != chResult.getErrodCode()) {
            logger.error(String.format("Reigister error, msg is %s", resultInfo.getErrorCodemsg()));
            return resultInfo.setData(false);
        }
        return resultInfo.setData(true);
    }

    /**
     * connect with vcenter server and unregister the plugin
     * @return
     */
    public ResultInfo unregisterPlugin(VCenterInfo vCenterInfo) {
        logger.info("-----------------Begin unregiser the plugin!-----------------");
        ResultInfo resultInfo = new ResultInfo();
        ErrorCode chResult = pluginMgntService.unRegister(vCenterInfo);
        resultInfo.setErrorCode(chResult.getErrodCode().intValue());
        resultInfo.setErrorDesc(chResult.getErrorDESC());
        if (ErrorCode.SUCCESS.getErrodCode() != chResult.getErrodCode()) {
            logger.error(String.format("Unreigister error, msg is %s", resultInfo.getErrorCodemsg()));
            return resultInfo.setData(false);
        }
        return resultInfo.setData(true);
    }

    /**
     * get vcenter Info from xml
     * @return
     */
    public VCenterInfo getvCenterInfo() {
        logger.info("-----------------In Home page!-----------------");
        VCenterInfo vcInfo = XmlUtils.getXml().xmlReadVcenterInfo();
        logger.info(String.format("Get vcenter info from local xml file! Info [%s].", vcInfo.toString()));
        return vcInfo;
    }

    /**
     * check the file is valid
     * @param fileName
     * @return
     */
    public boolean isValidZipFileName(String fileName) {
        return CommonUtils.isTheCorrectZipFile(fileName);
    }

    /**
     * return the ngc zip file path
     * @return
     */
    public String getZipfilePath() {
        return CommonUtils.getZipfilePath();
    }
}
