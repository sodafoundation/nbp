package org.opensds.vmware.ngc.service;


import org.opensds.vmware.ngc.entity.ResultInfo;

public interface SnapshotService {

    /**
     * get the snapshot count from one lun
     * @param VolumeId voulume id
     * @param storageId storage id
     * @return list of snapshot info
     */
    ResultInfo<Object> getSnapshotsCountByVolumeId(String VolumeId, String storageId);


    /**
     * get snapshot list from the volume
     * @param VolumeId volume id
     * @param storageId deviceId
     * @param start page start
     * @param count page count
     * @return
     */
    ResultInfo<Object> getSnapshotsByVolumeId(String VolumeId, String storageId, int start, int count);

    /**
     * create snapshot
     * @param snapshotName snapshot name
     * @param volumeId volume id
     * @param storageId storage id
     * @return boolen create snapshot success or failed
     */
    ResultInfo<Object> createVolumeSnapshot(String snapshotName, String volumeId, String storageId);

    /**
     * delete snapshot
     * @param snapshotIds array of snapshot
     * @param storageId storage id
     * @return boolen delete snapshot success or failed
     */
    ResultInfo<Object> deleteVolumeSnapshot(String[] snapshotIds, String storageId);

    /**
     * rollback snapshot of volume
     * @param snapshotId snapshot id
     * @param rollspeed the speed of roll back :
     *                  SPEED_LEVEL_LOW,
     *                  SPEED_LEVEL_MIDDLE,
     *                  SPEED_LEVEL_HIGH,
     *                  SPEED_LEVEL_ASAP
     * @param storageId storage id
     * @return
     */
    ResultInfo<Object> rollBackVolumeSnapShot(String snapshotId, String rollspeed, String storageId);

}
