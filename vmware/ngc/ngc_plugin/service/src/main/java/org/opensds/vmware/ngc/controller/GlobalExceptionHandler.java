package org.opensds.vmware.ngc.controller;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpServletResponse;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.util.HashMap;
import java.util.Map;


@ControllerAdvice
public class GlobalExceptionHandler {
    @ExceptionHandler(Exception.class)
    @ResponseBody
    public Map<String, String> handleException(Exception ex, HttpServletResponse response) {
        response.setStatus(HttpStatus.INTERNAL_SERVER_ERROR.value());

        Map<String, String> errorMap = new HashMap<String, String>();
        errorMap.put("message", ex.getMessage());
        if (ex.getCause() != null) {
            errorMap.put("cause", ex.getCause().getMessage());
        }
        StringWriter sw = new StringWriter();
        PrintWriter pw = new PrintWriter(sw);
        ex.printStackTrace(pw);
        errorMap.put("stackTrace", sw.toString());

        return errorMap;
    }
}
