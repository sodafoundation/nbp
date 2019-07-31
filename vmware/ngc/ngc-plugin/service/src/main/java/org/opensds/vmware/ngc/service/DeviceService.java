package org.opensds.vmware.ngc.service;

import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DeviceInfo;


public interface DeviceService {

    ResultInfo<Object> add(DeviceInfo device);

    ResultInfo<Object> update(Object deviceReference, DeviceInfo device);

    ResultInfo<Object> delete(Object deviceReference);

    ResultInfo<Object> get(Object deviceReference);

    ResultInfo<Object> getAllDeviceType();

    ResultInfo<Object> getList();

    ResultInfo<Object> getDeviceBlockPools(String deviceId);

}
