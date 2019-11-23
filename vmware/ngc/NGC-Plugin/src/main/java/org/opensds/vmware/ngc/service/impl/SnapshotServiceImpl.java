package org.opensds.vmware.ngc.service.impl;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.expections.ExpectionHandle;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.model.SnapshotInfo;
import org.opensds.vmware.ngc.models.SnapshotMO;
import org.opensds.vmware.ngc.service.SnapshotService;
import org.opensds.vmware.ngc.util.ListUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class SnapshotServiceImpl implements SnapshotService {

    private static final Log logger = LogFactory.getLog(SnapshotServiceImpl.class);

    private static final Map<String, List<SnapshotInfo>> CACHE_SNAPSHOTINFO = new ConcurrentHashMap<>();

    @Autowired
    private DeviceRepository deviceRepository;

    private static final String ERROR = "error";
    private static final String OK = "ok";

    /**
     * get the snapshot count from one lun
     * @param volumeId
     * @param storageId
     * @return list of snapshot info
     */
    @Override
    public ResultInfo<Object> getSnapshotsCountByVolumeId(
            String volumeId,
            String storageId) {
        logger.info(String.format(Locale.ROOT,"-----------Get the snapshot count from the volume(%s)!",volumeId));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        DeviceInfo deviceInfo = deviceRepository.get(storageId);
        Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);

        if (storage == null) {
            String strMsg = String.format(Locale.ROOT,"Can not find the storage by %s!", deviceInfo.uid);
            resultInfo.setMsg(strMsg);
            logger.error(strMsg);
            return resultInfo;
        }
        try {
            List<SnapshotMO> snapshotMOS = storage.listSnapshot(volumeId);
            List<SnapshotInfo> snapshotInfos = new ArrayList<>();
            snapshotMOS.forEach( n ->{
                SnapshotInfo snapshotInfo = new SnapshotInfo();
                snapshotInfo.convertSnapShotMo2Info(n).updateWithStorage(deviceInfo);
                snapshotInfos.add(snapshotInfo);
            });
            resultInfo.setData(snapshotInfos.size());
            CACHE_SNAPSHOTINFO.put(volumeId, snapshotInfos);
            logger.info(String.format(Locale.ROOT,"-----------Get the snapshot finished with count(%s).",
                    snapshotInfos.size()));
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        return resultInfo;
    }

    /**
     * get snapshot list from the volume
     * @param volumeId volume id
     * @param storageId deviceId
     * @param start page start
     * @param count page count
     * @return list of snapshots
     */
    @Override
    public ResultInfo<Object> getSnapshotsByVolumeId(
            String volumeId,
            String storageId,
            int start,
            int count) {
        logger.info(String.format(Locale.ROOT,"-----------Get the snapshot from the volume(%s)!",volumeId));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        List<SnapshotInfo> snapshotInfos = CACHE_SNAPSHOTINFO.get(volumeId);

        if (snapshotInfos != null && snapshotInfos.size() > 0) {
            resultInfo.setData(ListUtil.safeSubList(snapshotInfos, start, start + count));
        } else {
            String errorMsg = "Can not get the snapshots list!";
            logger.error(errorMsg);
            resultInfo.setMsg(errorMsg);
        }
        return resultInfo;
    }

    /**
     * create snapshot
     * @param snapshotName snapshot name
     * @param volumeId volume id
     * @param storageId storage id
     * @return boolen success or failed
     */
    @Override
    public ResultInfo<Object> createVolumeSnapshot(
            String snapshotName,
            String volumeId,
            String storageId) {
        logger.info(String.format(Locale.ROOT,"-----------Begin create snapshot from the volume(%s)!",
                volumeId));
        ResultInfo<Object> resultInfo = new ResultInfo<>();
        DeviceInfo deviceInfo = deviceRepository.get(storageId);
        Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);

        if (storage == null) {
            String strMsg = String.format(Locale.ROOT,"Can not find the storage by %s!", deviceInfo.uid);
            resultInfo.setMsg(strMsg);
            logger.error(strMsg);
            return resultInfo;
        }
        try {
            storage.createVolumeSnapshot(volumeId, snapshotName);
            resultInfo.setStatus(OK);
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        logger.info("-----------Create snapshot finished!");
        return resultInfo;
    }

    /**
     * delete snapshot
     * @param snapshotIds array of snapshot
     * @param storageId storage id
     * @return boolen delete snapshot success or failed
     */
    @Override
    public ResultInfo<Object> deleteVolumeSnapshot(
            String[] snapshotIds,
            String storageId) {
        logger.info(String.format(Locale.ROOT,"-----------Begin delelte snapshot!"));
        ResultInfo<Object> resultInfo = new ResultInfo<>();

        DeviceInfo deviceInfo = deviceRepository.get(storageId);
        Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);

        if (storage == null) {
            String strMsg = String.format(Locale.ROOT,"Can not find the storage by %s!", deviceInfo.uid);
            resultInfo.setMsg(strMsg);
            logger.error(strMsg);
            return resultInfo;
        }
        try {
            for (String id : snapshotIds) {
                storage.deleteVolumeSnapshot(id);
            }
            resultInfo.setStatus(OK);
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        logger.info(String.format(Locale.ROOT,"-----------Delelte snapshot finished!"));
        return resultInfo;
    }

    /**
     *
     * @param snapshotId snapshot id
     * @param rollspeed the speed of roll back :
     *                  SPEED_LEVEL_LOW,
     *                  SPEED_LEVEL_MIDDLE,
     *                  SPEED_LEVEL_HIGH,
     *                  SPEED_LEVEL_ASAP
     * @param storageId storage id
     * @return
     */
    @Override
    public ResultInfo<Object> rollBackVolumeSnapShot(
            String snapshotId,
            String rollspeed,
            String storageId){
        logger.info(String.format(Locale.ROOT,"-----------Begin rollback snapshot(%s)!", snapshotId));

        ResultInfo<Object> resultInfo = new ResultInfo<>();
        DeviceInfo deviceInfo = deviceRepository.get(storageId);
        Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);

        if (storage == null) {
            String strMsg = String.format(Locale.ROOT,"Can not find the storage by %s!", deviceInfo.uid);
            resultInfo.setMsg(strMsg);
            logger.error(strMsg);
            return resultInfo;
        }
        try {
            storage.rollbackVolumeSnapshot(snapshotId, rollspeed);
            resultInfo.setStatus(OK);
        } catch (Exception ex) {
            ExpectionHandle.handleExceptions(resultInfo, ex);
        }
        logger.info(String.format(Locale.ROOT,"-----------Rollback snapshot finished!", snapshotId));
        return resultInfo;
    }

}
