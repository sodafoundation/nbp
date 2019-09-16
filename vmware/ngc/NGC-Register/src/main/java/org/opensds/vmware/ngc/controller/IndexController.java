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

package org.opensds.vmware.ngc.controller;

import org.opensds.vmware.ngc.bean.ErrorCode;
import org.opensds.vmware.ngc.model.ResultInfo;
import org.opensds.vmware.ngc.model.VCenterInfo;
import org.opensds.vmware.ngc.service.VCenterService;
import org.opensds.vmware.ngc.utils.FileUtils;
import org.opensds.vmware.ngc.utils.MediaTypeUtils;
import org.opensds.vmware.ngc.utils.OperationSystemUtils;
import org.opensds.vmware.ngc.utils.XmlUtils;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.core.io.InputStreamResource;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.validation.BindingResult;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.context.ServletContextAware;

import javax.servlet.ServletContext;
import javax.validation.Valid;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;


@Controller
public class IndexController implements  ServletContextAware {

    @Autowired
    private VCenterService vCenterService;

    private ServletContext servletContext;

    private static final Logger logger = LogManager.getLogger(IndexController.class);

    private static final String ERROR_PAGE = "error";

    private static final String SUCCESS_PAGE = "success";

    private static final String RESULT_INFO = "resultInfo";

    private static final String Home_Page = "homePage";

    public void setServletContext(ServletContext servletContext) {
        this.servletContext = servletContext;
    }

    /**
     * direct to home page
     * @param model
     * @return
     */
    @RequestMapping(value = "/homePage")
    public String home (Model model) {
        VCenterInfo vcInfo = vCenterService.getvCenterInfo();
        model.addAttribute("VCenterInfo", vcInfo);
        return Home_Page;
    }

    /**
     * direct to register action
     * @param vcInfo
     * @param bindingResult
     * @return
     */
    @RequestMapping(value = "/action", method = RequestMethod.POST, params = "action=register")
    public String registerSubmit (@Valid @ModelAttribute VCenterInfo vcInfo,
                                 BindingResult bindingResult,
                                 Model model) {

        if (bindingResult.hasErrors()) {
            model.addAttribute(RESULT_INFO, new ResultInfo(ErrorCode.CONNECT_FAIL.getErrodCode().intValue(),
                    ErrorCode.CONNECT_FAIL.getErrorDESC()));
            return ERROR_PAGE;
        }
        ResultInfo result = vCenterService.registerPlugin(vcInfo);
        if ((boolean)result.getData() == true) {
            model.addAttribute(RESULT_INFO, "Register success!");
            return SUCCESS_PAGE;
        } else {
            model.addAttribute(RESULT_INFO, result.getErrorDesc());
            return ERROR_PAGE;
        }
    }

    /**
     * direct to unregister aciton
     * @return
     */
    @RequestMapping(value = "/action", method = RequestMethod.POST, params = "action=unregister")
    public String doUnregister (@Valid @ModelAttribute VCenterInfo vcInfo,
                                 BindingResult bindingResult,
                                 Model model) {

        if (bindingResult.hasErrors()) {
            model.addAttribute(RESULT_INFO, new ResultInfo(ErrorCode.CONNECT_FAIL.getErrodCode().intValue(),
                    ErrorCode.CONNECT_FAIL.getErrorDESC()));
            return ERROR_PAGE;
        }
        ResultInfo result = vCenterService.unregisterPlugin(vcInfo);
        if ((boolean)result.getData() == true) {
            model.addAttribute(RESULT_INFO, "Unregister success!");
            return SUCCESS_PAGE;
        } else {
            model.addAttribute(RESULT_INFO, result.getErrorDesc());
            return ERROR_PAGE;
        }
    }

    /**
     * get the ngc zip file
     * @param fileName
     * @return
     * @throws IOException
     */
    @RequestMapping(value="/download/{fileName}", method= RequestMethod.GET)
    public ResponseEntity<InputStreamResource> downloadFile (@PathVariable("fileName") String fileName)
            throws IOException {
        logger.info(String.format("-----------------Begin download the zip file %s-----------------", fileName));
        if (!vCenterService.isValidZipFileName(fileName)) {
            return ResponseEntity.notFound().build();
        }
        MediaType mediaType = MediaTypeUtils.getMediaTypeForFileName(this.servletContext, fileName);
        File file = new File(vCenterService.getZipfilePath());
        InputStreamResource resource = new InputStreamResource(new FileInputStream(file));

        return ResponseEntity.ok()
                .header(HttpHeaders.CONTENT_DISPOSITION, "attachment;filename=" + file.getName())
                // Content-Type
                .contentType(mediaType)
                // Contet-Length
                .contentLength(file.length())
                .body(resource);
    }
}
