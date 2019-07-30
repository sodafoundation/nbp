package org.opensds.vmware.ngc.controller;

import com.google.gson.Gson;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DatastoreInfo;
import org.opensds.vmware.ngc.service.DatastoreService;
import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;

@Controller
@RequestMapping(value = "/datastore")
public class DatastoreController {

    private static final Log logger = LogFactory.getLog(DatastoreController.class);

    @Autowired
    private DatastoreService datastoreService;


    @RequestMapping(value = "/createDatastores", method = RequestMethod.POST)
    @ResponseBody
    public ResultInfo createDatastore(
            @RequestParam(value = "objectId") ManagedObjectReference[] hostMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo,
            @RequestParam(value = "json") String json)
            throws Exception {
        Gson gson = new Gson();
        DatastoreInfo datastoreInfo = gson.fromJson(json, DatastoreInfo.class);
        ResultInfo resultInfo = datastoreService.createDatastore(hostMoRef, serverInfo, datastoreInfo);
        return resultInfo;
    }
}
