package org.opensds.vmware.ngc.dao;

import org.opensds.vmware.common.Storage;
import org.opensds.vmware.ngc.model.DeviceInfo;


import java.util.Map;

public interface DeviceRepository {


    Map<String, DeviceInfo> getAll();

    void remove(String uid);

    void update(String uid, DeviceInfo device) throws Exception;

    void add(String uid , DeviceInfo device) throws Exception;

    DeviceInfo get(String uid);

    Storage getLoginedDeviceByIP(String deviceIP);


}
