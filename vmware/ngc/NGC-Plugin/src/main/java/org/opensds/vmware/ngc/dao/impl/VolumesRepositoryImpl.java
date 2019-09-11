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

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.dao.VolumesRepository;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Lazy;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;


@Repository
@Lazy(value = false)
public class VolumesRepositoryImpl implements VolumesRepository {

    private static final Log logger = LogFactory.getLog(VolumesRepository.class);

    // key : deviceId
    // value : list of volumes wwn
    private static Map<String, List<String>> CACHE_VOLUMES_INFO = new ConcurrentHashMap<>();

    @Autowired
    private DeviceRepository deviceRepository;


    /**
     * add the device with volumes in backend;
     *
     * @param deviceInfo device
     */
    public void addDevice(DeviceInfo deviceInfo) {
        ExecutorService executorService = Executors.newCachedThreadPool();
        executorService.submit((Runnable) () -> {
            try {
                logger.info("-----------Begin add device volumes!");
                Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
                if (storage == null) {
                    logger.error("Can not find the device!");
                    return;
                }
                List<String> listWwns = new ArrayList();
                storage.listVolumes().forEach(n -> listWwns.add(n.wwn));
                update(deviceInfo.uid, listWwns);
            } catch (Exception ex) {
                logger.error(ex.getMessage());
            }
        });
        executorService.shutdown();
    }

    /**
     * remove the deveice
     *
     * @param deviceId device id
     */
    @Override
    public void removeDevice(String deviceId) {
        CACHE_VOLUMES_INFO.remove(deviceId);
    }

    /**
     * update the device and volume id
     *
     * @param deviceID   device id
     * @param volumeWWNs list of volumes
     */
    @Override
    public synchronized void update(String deviceID, List<String> volumeWWNs) {
        logger.info("Update the Volumes with wwn in Repository");
        if (CACHE_VOLUMES_INFO.containsKey(deviceID)) {
            List<String> backTmp = CACHE_VOLUMES_INFO.get(deviceID);
            if (!backTmp.contains(volumeWWNs)) {
                CACHE_VOLUMES_INFO.get(deviceID).addAll(volumeWWNs);
            }
        } else {
            CACHE_VOLUMES_INFO.put(deviceID, volumeWWNs);
        }
    }

    /**
     * reomve one volume wwn
     *
     * @param wwn String wwn
     */
    @Override
    public synchronized void removeVolume(String deviceID, String wwn) {
        if (CACHE_VOLUMES_INFO.containsKey(deviceID)) {
            CACHE_VOLUMES_INFO.get(deviceID).remove(wwn);
        }
    }

    /**
     * get Device info by wwn
     *
     * @param wwn
     * @return
     */
    @Override
    public DeviceInfo getDevicebyWWN(String wwn) throws Exception {
        logger.info(String.format(Locale.ROOT, "Begin get device by wwn(%s) in cache!", wwn));
        // test
        for (String uid : CACHE_VOLUMES_INFO.keySet()) {
            logger.info("String key: " + uid);
        }
        //

        for (Object key : CACHE_VOLUMES_INFO.keySet()) {
            if (CACHE_VOLUMES_INFO.get(key).contains(wwn)) {
                return deviceRepository.get((String) key);
            }
        }
        return getDeviceInRepository(wwn);
    }

    // can not found in the cache, get device info in repository, update the cache
    private DeviceInfo getDeviceInRepository(String volumeWWN) throws Exception {
        logger.info("Can not find in cache, begin query device in repository...");
        for (DeviceInfo deviceInfo : deviceRepository.getAll().values()) {
            Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
            if (storage == null) {
                logger.error("Can not get the device :" + deviceInfo.ip);
                continue;
            }
            logger.info(".... Get the device ip :" + deviceInfo.ip);
            try {
                VolumeMO volumeMO = storage.queryVolumeByID(volumeWWN);
                if (volumeMO != null) {
                    List<String> tmp = new ArrayList<>();
                    tmp.add(volumeWWN);
                    update(deviceInfo.uid, tmp);
                    logger.info(String.format(Locale.ROOT, "Get volume(%s) in device(%s) ", volumeWWN, deviceInfo.ip));
                    return deviceInfo;
                }
            } catch (Exception ex) {
                logger.info("Can not query the voulume :" + ex.getMessage());
            }
        }
        throw new Exception(String.format("Can not find this volume(%s) in any devices!", volumeWWN));
    }
}
