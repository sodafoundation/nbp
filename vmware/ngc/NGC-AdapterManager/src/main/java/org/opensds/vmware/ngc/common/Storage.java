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

package org.opensds.vmware.ngc.common;

import org.opensds.vmware.ngc.models.*;

import java.util.List;

public abstract class Storage {

    protected String name;

    public Storage(String name) {
        this.name = name;
    }

    public abstract void login(String ip, int port, String user, String password) throws Exception;

    public abstract void logout();

    public abstract StorageMO getDeviceInfo() throws Exception;

    public abstract VolumeMO createVolume(String name, String description, ALLOC_TYPE allocType, long capacity, String poolId) throws
            Exception;

    public abstract void deleteVolume(String volumeId) throws Exception;

    public abstract List<VolumeMO> listVolumes() throws Exception;

    public abstract List<VolumeMO> listVolumes(String filterKey, String filtervalue) throws Exception;

    public abstract List<VolumeMO> listVolumes(String poolId) throws Exception;

    public abstract List<StoragePoolMO> listStoragePools() throws Exception;

    public abstract StoragePoolMO getStoragePool(String poolId) throws Exception;

    public abstract void attachVolume(String volumeId, ConnectMO connect) throws Exception;

    public abstract void detachVolume(String volumeId, ConnectMO connect) throws Exception;

    /**
     * query the volume mob by volume id
     *
     * @param volumeId in oceanstor volume id is wwn
     *                 in opensds volume id is volume id
     * @return volume mob
     * @throws Exception
     */
    public abstract VolumeMO queryVolumeByID(String volumeId) throws Exception;

    /**
     * get the list of volumes from a specified volume by volumeID
     *
     * @param volumeId the id of volume
     * @return list of volume
     *  @throws Exception
     */
    public abstract List<SnapshotMO> listSnapshot(String volumeId) throws Exception;

    /**
     * create the snapshot from a specified volume by volumeID
     * @param volumeId volume id
     * @param name volume name
     * @throws Exception
     */
    public abstract void createVolumeSnapshot(String volumeId, String name) throws Exception;

    /**
     * delete a specified snapshot by id
     * @param snapshotId snapshot id
     * @throws Exception
     */
    public abstract void deleteVolumeSnapshot(String snapshotId) throws Exception;

    /**
     * rollback a snapshot of the volume
     * @param snapshotId snapshot id
     * @param rollbackSpeed the rollback speed of the volume
     *                  UNKNOWN(-1),
     *                  SPEED_LEVEL_LOW (1),
     *                  SPEED_LEVEL_MIDDLE (2),
     *                  SPEED_LEVEL_HIGH (3),
     *                  SPEED_LEVEL_ASAP (4)
     *
     * @throws Exception
     */
    public abstract void rollbackVolumeSnapshot(String snapshotId, String rollbackSpeed) throws Exception;
	
	 /**
     * extend the volume size
     * @param volumeId volume
     * @param capacity
     */
    public abstract void expandVolume(String volumeId, long capacity) throws Exception;
}
