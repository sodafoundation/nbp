package org.opensds.vmware.ngc.expections;


import org.opensds.vmware.ngc.entity.ResultInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

public class ExpectionHandle {

    private static Log _logger = LogFactory.getLog(ExpectionHandle.class);

    public static void handleExceptions(ResultInfo resultInfo, Throwable e) {
        _logger.error(e.getMessage());
        resultInfo.setMsg(e.getMessage());
        if (e instanceof InactiveSessionException){
            resultInfo.setStatus("error");
        }
    }
}
