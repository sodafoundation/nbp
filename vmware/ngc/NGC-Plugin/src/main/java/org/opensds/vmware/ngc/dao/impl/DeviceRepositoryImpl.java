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

package org.opensds.vmware.ngc.dao.impl;

import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.common.StorageFactory;
import org.opensds.vmware.ngc.models.StorageMO;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.base.Constant;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.context.annotation.Lazy;
import org.springframework.stereotype.Repository;

import javax.annotation.PostConstruct;
import java.io.ObjectInputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.Closeable;
import java.io.ObjectOutputStream;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

@Repository
@Lazy(value = false)
public class DeviceRepositoryImpl implements DeviceRepository {

    private static final Log logger = LogFactory.getLog(DeviceRepositoryImpl.class);
    private static Map<String, DeviceInfo> CACHE_DEVICE_INFO = new ConcurrentHashMap<>();
    private static Map<String, Storage> LOGINED_DEVICE = new ConcurrentHashMap<>();
    private static final String DEVICE_DATA_FILE_PATH = Constant.DEVICE_DATA_FILE;

    @PostConstruct
    private void init() {
        ObjectInputStream objectInputStream = null;
        try {
            File file = new File(DEVICE_DATA_FILE_PATH);
            if (!file.exists()) {
                return;
            }
            objectInputStream = new ObjectInputStream(new FileInputStream(file));
            Map<String, DeviceInfo> deviceInfoMap = (Map<String, DeviceInfo>) objectInputStream.readObject();
            reLogin(deviceInfoMap);
            CACHE_DEVICE_INFO.putAll(deviceInfoMap);
        } catch (IOException e) {
            logger.warn(e.getMessage(), e);
        } catch (ClassNotFoundException e) {
            logger.warn(e.getMessage(), e);
        } finally {
            quietClose(objectInputStream);
        }
    }

    @Override
    public Map<String, DeviceInfo> getAll() {
        return CACHE_DEVICE_INFO;
    }

    @Override
    public void remove(String uid) {
        DeviceInfo deviceInfo = CACHE_DEVICE_INFO.remove(uid);
        LOGINED_DEVICE.remove(deviceInfo.uid);
        writeDataToFile();
    }

    @Override
    public void update(String uid, DeviceInfo deviceInfo) throws Exception {
        initDevice(deviceInfo, uid);
        CACHE_DEVICE_INFO.put(uid, deviceInfo);
        writeDataToFile();
    }

    @Override
    public void add(String uid, DeviceInfo deviceInfo) throws Exception {
        initDevice(deviceInfo, uid);
        CACHE_DEVICE_INFO.put(uid, deviceInfo);
        writeDataToFile();
    }

    @Override
    public DeviceInfo get(String uid) {
        logger.info("uid :" +  uid);
        return CACHE_DEVICE_INFO.get(uid);
    }

    @Override
    public Storage getLoginedDeviceByID(String deviceUID) {
        return LOGINED_DEVICE.get(deviceUID);
    }


    private void writeDataToFile() {
        ObjectOutputStream objectOutputStream = null;
        try {
            File file = new File(DEVICE_DATA_FILE_PATH);
            objectOutputStream = new ObjectOutputStream(new FileOutputStream(file));
            synchronized (CACHE_DEVICE_INFO) {
                objectOutputStream.writeObject(CACHE_DEVICE_INFO);
            }
        } catch (IOException e) {
            logger.error(e.getMessage(), e);
        } finally {
            quietClose(objectOutputStream);
        }
    }

    private void initDevice(DeviceInfo deviceInfo, String uid) throws Exception {
        Storage device = StorageFactory.newStorage(getDeviceClassName(deviceInfo.deviceType), "");
        device.login(deviceInfo.ip, deviceInfo.port, deviceInfo.username, deviceInfo.password);
        StorageMO stroageMO = device.getDeviceInfo();
        deviceInfo.deviceModel = stroageMO.model;
        deviceInfo.deviceName = stroageMO.name;
        deviceInfo.deviceStatus = stroageMO.status;
        deviceInfo.sn = stroageMO.sn;
        deviceInfo.uid = uid;
        LOGINED_DEVICE.put(deviceInfo.uid, device);
    }

    private String getDeviceClassName(String deviceType) {
        String[] types = StorageFactory.listStorages();
        for (String type : types) {
            if (type.substring(type.lastIndexOf(".") + 1).equals(deviceType)) {
                return type;
            }
        }
        return "";
    }

    private void quietClose(Closeable closeable) {
        if (closeable != null) {
            try {
                closeable.close();
            } catch (IOException e) {
                logger.error(e.getMessage());
            }
        }
    }

    private void reLogin(Map<String, DeviceInfo> deviceInfoMap) {
        if (deviceInfoMap == null || deviceInfoMap.isEmpty()) {
            return;
        }
        ExecutorService executorService = Executors.newCachedThreadPool();
        int deviceNumber = deviceInfoMap.size();
        final CountDownLatch countDownLatch = new CountDownLatch(deviceNumber);
        for (final Map.Entry<String, DeviceInfo> entry : deviceInfoMap.entrySet()) {
            executorService.submit(new Runnable() {
                @Override
                public void run() {
                    try {
                        DeviceRepositoryImpl.this.initDevice(entry.getValue(), entry.getKey());
                    } catch (Exception e) {
                        logger.error("The device " + entry.getValue().ip + " fail to login");
                    } finally {
                        countDownLatch.countDown();
                    }
                }
            });
        }
        try {
            countDownLatch.await(5, TimeUnit.MINUTES);
        } catch (InterruptedException e) {
            logger.error(e.getMessage());
        }
        executorService.shutdown();
    }
}
