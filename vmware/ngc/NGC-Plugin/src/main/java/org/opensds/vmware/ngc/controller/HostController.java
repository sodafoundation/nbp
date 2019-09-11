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

package org.opensds.vmware.ngc.controller;


import com.vmware.vim25.ManagedObjectReference;
import com.vmware.vise.usersession.ServerInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.opensds.vmware.ngc.base.VimFieldsConst;
import org.opensds.vmware.ngc.common.Storage;
import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.entity.ResultInfo;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.model.SnapshotInfo;
import org.opensds.vmware.ngc.service.DataCenterService;
import org.opensds.vmware.ngc.service.HostService;
import org.opensds.vmware.ngc.service.SnapshotService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;

/**
 * Query exi host info
 */
@Controller
@RequestMapping(value = "/data/host")
public class HostController {

    private static final Log logger = LogFactory.getLog(HostController.class);

    @Autowired
    private HostService hostService;

    @Autowired
    private DataCenterService dataCenterService;

    @Autowired
    private DeviceRepository deviceRepository;

    @Autowired
    private SnapshotService snapshotService;

    /**
     * get esxi host list info
     *
     * @param ObjectMOR  ManagedObjectReference
     * @param serverInfo ServerInfo
     * @return ResultInfo<Object>
     */
    @RequestMapping(value = "/getESXIList", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getHostList(
            @RequestParam(value = "objectId") ManagedObjectReference ObjectMOR,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        String type = ObjectMOR.getType();
        if (type.equalsIgnoreCase(VimFieldsConst.MoTypesConst.Datacenter)) {
            return dataCenterService.getHostListByDataCenterId(ObjectMOR, serverInfo);
        }
        if (type.equalsIgnoreCase(VimFieldsConst.MoTypesConst.ClusterComputeResource)) {
            return dataCenterService.getHostListByClusterId(ObjectMOR, serverInfo);
        }
        logger.error(" input is not datacenterName or clusterName!");
        return new ResultInfo();
    }

    /**
     * get esxi host status
     *
     * @param hostMoRef  ManagedObjectReference
     * @param serverInfo ServerInfo
     * @return ResultInfo<Object>
     */
    @RequestMapping(value = "/getEXIStatus/{hostId}", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getHostStatus(
            @PathVariable("hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getHostConnectionState(hostMoRef, serverInfo);
    }

    /**
     * get the datastores belongs to this ESXI
     *
     * @param hostMoRef  hostSystem
     * @param serverInfo ServerInfo
     * @return list of datastore
     */
    @RequestMapping(value = "/datastores/{hostId}", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getDatastoreList(
            @PathVariable("hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getDatastoreList(hostMoRef, serverInfo);
    }

    /**
     * Get volumes from the datastore
     *
     * @param datastoreID
     * @param hostMoRef
     * @param serverInfo
     * @return list of volumes
     */
    @RequestMapping(value = "/volumesOfdatastore/{datastoreID}", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getVolumesfromDatastore(
            @PathVariable("datastoreID") String datastoreID,
            @RequestParam(value = "hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getVolumesFromDatastore(datastoreID, hostMoRef, serverInfo);
    }


    /**
     * Get volumes which attach to host
     * @param hostMo host mob
     * @param serverInfo service mob
     * @return list of volumes
     */
    @RequestMapping(value = "/volumeListOfHost/{hostId}", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getVolumesofHost(
            @PathVariable("hostId") ManagedObjectReference hostMo,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getVolumeofHost(hostMo, serverInfo);
    }

    /**
     * get the luns without mount with the ESXI
     *
     * @param hostMoRef   host mob
     * @param deviceId    storage id
     * @param filterType  filter type
     * @param filterValue filter value
     * @param serverInfo  server info
     * @return num of luns
     */
    @RequestMapping(value = "/mountableVolumeList/count/{hostId}",method = {RequestMethod.GET,RequestMethod.POST})
    @ResponseBody
    public ResultInfo<Object> getMountableVolumeListCount(
            @PathVariable("hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "deviceId") String deviceId,
            @RequestParam(value = "filterType") String filterType,
            @RequestParam(value = "filterValue") String filterValue,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getMountableVolumeListCount(hostMoRef, serverInfo, deviceId, filterType, filterValue);
    }

    /**
     * get mountable volume list of the ESXI
     *
     * @param hostMoRef   host mob
     * @param deviceId    device id
     * @param filterType  filterType
     * @param filterValue filterValue
     * @param start       page start
     * @param count       page count
     * @param serverInfo  sever instance
     * @return list of voumluie
     */
    @RequestMapping(value = "/mountableVolumeList/{hostId}", method = {RequestMethod.GET,RequestMethod.POST})
    @ResponseBody
    public ResultInfo<Object> getMountableVolumeList(
            @PathVariable("hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "deviceId") String deviceId,
            @RequestParam(value = "filterType") String filterType,
            @RequestParam(value = "filterValue") String filterValue,
            @RequestParam(value = "start") int start,
            @RequestParam(value = "count") int count,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getMountableVolumeList(hostMoRef, serverInfo, deviceId, filterType, filterValue, start,
                count);
    }

    /**
     * mount free volumes to the esxi
     *
     * @param hostMo
     * @param deviceIds
     * @param volumeIds
     * @param serverInfo
     * @return mount volumes success of failed
     */
    @RequestMapping(value = "/mountVolumes/{hostId}", method = RequestMethod.PUT)
    @ResponseBody
    public ResultInfo<Object> mountFreeVolumes(
            @PathVariable("hostId") ManagedObjectReference hostMo,
            @RequestParam(value = "deviceIds") String deviceIds,
            @RequestParam(value = "volumeIds") String volumeIds,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {

        logger.info("---------------Mount the free volumes to the esxi!");
        String deviceId = deviceIds.split(",")[0];
        DeviceInfo deviceInfo = deviceRepository.get(deviceId);
        Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
        String[] volumeIdList = volumeIds.split(",");
        ManagedObjectReference[] hostMos = new ManagedObjectReference[1];
        hostMos[0] = hostMo;
        ResultInfo<Object> resultInfo = hostService.mountVolumesByIds(hostMos, serverInfo, storage, volumeIdList);
        if (resultInfo.getMsg() != null) {
            return resultInfo;
        }
        return hostService.rescanAllHba(hostMo, serverInfo);
    }

    /**
     * Get the count of unmount the volume list from the the host
     *
     * @param hostMoRef   esix host mob
     * @param filterType  filter type string
     * @param filterValue filter type value
     * @param serverInfo  server instance
     * @return count of the unable volume list
     */
    @RequestMapping(value = "/unmountableVolumeList/count/{hostId}", method = {RequestMethod.GET,RequestMethod.POST})
    @ResponseBody
    public ResultInfo<Object> getUnmountableVolumeListCount(
            @PathVariable("hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "filterType") String filterType,
            @RequestParam(value = "filterValue") String filterValue,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getUnmountableVolumeListCount(hostMoRef, serverInfo, filterType, filterValue);
    }

    /**
     * get unmountable volume list of the ESXI
     *
     * @param hostMoRef   host mob
     * @param filterType  filterType
     * @param filterValue filterValue
     * @param start       page start
     * @param count       page count
     * @param serverInfo  sever instance
     * @return list of volumes
     */
    @RequestMapping(value = "/unmountableVolumeList/{hostId}", method = {RequestMethod.GET,RequestMethod.POST})
    @ResponseBody
    public ResultInfo<Object> getUnmountableVolumes(
            @PathVariable("hostId") ManagedObjectReference hostMoRef,
            @RequestParam String filterType,
            @RequestParam String filterValue,
            @RequestParam(value = "start") int start,
            @RequestParam(value = "count") int count,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        return hostService.getUnmountableVolumes(hostMoRef, serverInfo, filterType, filterValue, start, count);
    }

    /**
     * unmount volumes from the esxi
     *
     * @param hostMo
     * @param deviceIds
     * @param volumeIds
     * @param serverInfo
     * @returnmount unmount volumes success of failed
     */
    @RequestMapping(value = "/unmountVolumes/{hostId}", method = RequestMethod.PUT)
    @ResponseBody
    public ResultInfo<Object> unmountVolumes(
            @PathVariable("hostId") ManagedObjectReference hostMo,
            @RequestParam(value = "deviceIds") String deviceIds,
            @RequestParam(value = "volumeIds") String volumeIds,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        logger.info("---------------Unmount the volumes form the esxi!");
        String[] volumeIDArray = volumeIds.split(",");
        String[] deviceIdArray = deviceIds.split(",");

        for (int i = 0; i < deviceIdArray.length; i++) {
            DeviceInfo deviceInfo = deviceRepository.get(deviceIdArray[i]);
            Storage storage = deviceRepository.getLoginedDeviceByID(deviceInfo.uid);
            ResultInfo<Object> resultInfo = hostService.unmountVolume(hostMo, serverInfo, storage, volumeIDArray[i]);
            if (resultInfo.getMsg() != null) {
                return resultInfo;
            }
        }
        return hostService.rescanAllHba(hostMo, serverInfo);
    }


    /**
     * get Snapshot count from the volume
     *
     * @param volumeId  volume id
     * @param storageId storage id
     * @return count of snapshots
     */
    @RequestMapping(value = "/snapshot/count", method = {RequestMethod.GET, RequestMethod.POST})
    @ResponseBody
    public ResultInfo<Object> getSanpshotsCountByVolumeId(
            @RequestParam(value = "volumeId") String volumeId,
            @RequestParam(value = "storageId") String storageId) {
        ResultInfo<Object> resultInfo = snapshotService.getSnapshotsCountByVolumeId(volumeId, storageId);
        return resultInfo;
    }

    /**
     * get snapshot list from the volume
     *
     * @param lunId    volume id
     * @param deviceId deviceId
     * @param start    page start
     * @param count    page count
     * @return list of snapshots
     */
    @RequestMapping(value = "/snapshot", method = RequestMethod.GET)
    @ResponseBody
    public ResultInfo<Object> getSnapshotsByVolumeId(
            @RequestParam(value = "volumeId") String lunId,
            @RequestParam(value = "storageId") String deviceId,
            @RequestParam(value = "start") int start,
            @RequestParam(value = "count") int count) {
        return snapshotService.getSnapshotsByVolumeId(lunId, deviceId, start, count);
    }

    /**
     * create the snapshot by storage Id
     *
     * @param storageId     storage id
     * @param snapshotInfos snapshot infos
     * @param hostMoRef     host mob
     * @param serverInfo    server instance
     * @return
     */
    @RequestMapping(value = "/snapshot/{storageId:.+}", method = RequestMethod.DELETE)
    @ResponseBody
    public ResultInfo<Object> deleteVolumeSnashot(
            @PathVariable("storageId") String storageId,
            @RequestBody SnapshotInfo[] snapshotInfos,
            @RequestParam(value = "hostId") ManagedObjectReference hostMoRef,
            @RequestParam(value = "serverGuid") ServerInfo serverInfo) {
        logger.info("---------------Delete the snapshot id, snapshotInfos size:" + snapshotInfos.length);
        logger.info("-------storageId : " + storageId);
        String[] snapshots = new String[snapshotInfos.length];
        for (int i = 0; i < snapshotInfos.length; i++) {
            snapshots[i] = snapshotInfos[i].getId();
        }
        return snapshotService.deleteVolumeSnapshot(snapshots, storageId);
    }

    /**
     * create the snapshot
     *
     * @param snapshot snapshot info
     * @return boolen create the snapshot successful or failed
     */
    @RequestMapping(value = "/snapshot", method = RequestMethod.POST)
    @ResponseBody
    public ResultInfo<Object> createVolumeSnashot(
            @RequestBody SnapshotInfo snapshot) {
        logger.info("---------------create the snapshot id :" + snapshot.getId());
        return snapshotService.createVolumeSnapshot(snapshot.getName(), snapshot.getParentId(), snapshot.getStorageId());
    }

    /**
     *
     * @param storageId
     * @param rollbackSpeed
     * @param snapshotInfo snapshot info
     * @return
     */
    @RequestMapping(value = "/snapshot/{storageId:.+}", method = RequestMethod.PUT)
    @ResponseBody
    public ResultInfo<Object> rollBackVolumeSnapShot(
            @PathVariable("storageId") String storageId,
            @RequestParam("rollbackSpeed") String rollbackSpeed,
            @RequestBody SnapshotInfo snapshotInfo) {
        return snapshotService.rollBackVolumeSnapShot(snapshotInfo.getId(), rollbackSpeed, storageId);
    }

}
