package org.opensds.vmware.ngc.model;

import com.vmware.vim25.HostScsiDisk;
import com.vmware.vim25.ScsiLun;
import org.opensds.vmware.ngc.models.StoragePoolMO;
import org.opensds.vmware.ngc.models.VolumeMO;
import org.opensds.vmware.ngc.util.CapacityUtil;

public class VolumeInfo {


    public static final String UNKNOW = "";

    public static final String SZERO = "0";

    public static final double DZERO = 0;

    public static final Long LZERO = 0L;

    private String name;

    private String id;

    private String description;

    private String capacity;

    private String freeCapacity;

    private double capacityUsage;

    private String allocType;

    private String wwn;

    private Long extendSize;

    private String storagePoolName;

    private String storagePoolId;

    private String storageId;

    private String storagePoolCapactiy;

    private String storagePoolFreeCap;

    private double storagePoolUsage;

    private String storageType;

    private String storageIP;

    private boolean isLocal;

    private String identifier;

    private String status;

    private String usedBy;

    private String usedType;

    public VolumeInfo() {
        this.name = UNKNOW;
        this.id = UNKNOW;
        this.description = UNKNOW;
        this.capacity = SZERO;
        this.freeCapacity = SZERO;
        this.capacityUsage = DZERO;
        this.allocType = UNKNOW;
        this.wwn = UNKNOW;
        this.extendSize = LZERO;
        this.storagePoolName = UNKNOW;
        this.storagePoolId = UNKNOW;
        this.storageId = UNKNOW;
        this.storagePoolCapactiy = SZERO;
        this.storagePoolFreeCap = SZERO;
        this.storagePoolUsage = DZERO;
        this.storageType = UNKNOW;
        this.storageIP = UNKNOW;
        this.isLocal = false;
        this.identifier = UNKNOW;
        this.status = UNKNOW;
        this.usedBy = UNKNOW;
        this.usedType = UNKNOW;
    }


    public String getName() {
        return name;
    }

    public String getId() {
        return id;
    }

    public String getDescription() {
        return description;
    }

    public String getStoragePoolId() {
        return storagePoolId;
    }

    public String getStorageId() {
        return storageId;
    }

    public String getCapacity() {
        return capacity;
    }

    public String getFreeCapacity() {
        return freeCapacity;
    }

    public String getAllocType() {
        return allocType;
    }

    public void setName(String name) {
        this.name = name;
    }

    public void setId(String id) {
        this.id = id;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public void setStoragePoolId(String storagePoolId) {
        this.storagePoolId = storagePoolId;
    }

    public void setStorageId(String storageId) {
        this.storageId = storageId;
    }

    public void setFreeCapacity(String freeCapacity) {
        this.freeCapacity = freeCapacity;
    }

    public void setStoragePoolName(String storagePoolName) {
        this.storagePoolName = storagePoolName;
    }

    public void setAllocType(String allocType) {
        this.allocType = allocType;
    }

    public String getStorageType() {
        return storageType;
    }

    public void setStorageType(String storageType) {
        this.storageType = storageType;
    }

    public String getStoragePoolFreeCap() {
        return storagePoolFreeCap;
    }

    public void setStoragePoolFreeCap(String storagePoolFreeCap) {
        this.storagePoolFreeCap = storagePoolFreeCap;
    }

    public String getStoragePoolName() {
        return storagePoolName;
    }

    public String getWwn() {
        return wwn;
    }

    public void setWwn(String wwn) {
        this.wwn = wwn;
    }

    public String getStoragePoolCapactiy() {
        return storagePoolCapactiy;
    }

    public void setStoragePoolCapactiy(String storagePoolCapactiy) {
        this.storagePoolCapactiy = storagePoolCapactiy;
    }

    public Long getExtendSize() {
        return extendSize;
    }

    public void setExtendSize(Long extendSize) {
        this.extendSize = extendSize;
    }

    public String getStorageIP() {
        return storageIP;
    }

    public void setStorageIP(String storageIP) {
        this.storageIP = storageIP;
    }

    public boolean isLocal() {
        return isLocal;
    }

    public void setLocal(boolean local) {
        isLocal = local;
    }

    public String getIdentifier() {
        return identifier;
    }

    public void setIdentifier(String identifier) {
        this.identifier = identifier;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getUsedBy() {
        return usedBy;
    }

    public void setUsedBy(String usedBy) {
        this.usedBy = usedBy;
    }

    public void setUsedType(String usedType) {
        this.usedType = usedType;
    }

    public String getUsedType(){
        return usedType;
    }

    public double getStoragePoolUsage() {
        return storagePoolUsage;
    }

    public void setStoragePoolUsage(double storagePoolUsage) {
        this.storagePoolUsage = storagePoolUsage;
    }

    public double getCapacityUsage() {
        return capacityUsage;
    }

    public void setCapacityUsage(double capacityUsage) {
        this.capacityUsage = capacityUsage;
    }

    public void setCapacity(String capacity) {
        this.capacity = capacity;
    }

    /**
     * format volumeMo to VolueInfo
     *
     * @param volumeMO
     */
    public VolumeInfo convertVolumeMO2Info(VolumeMO volumeMO) {
        this.name = volumeMO.name;
        this.wwn = volumeMO.wwn;
        this.id = volumeMO.id;
        this.capacity = CapacityUtil.convert512BToCap(volumeMO.capacity);
        this.freeCapacity = volumeMO.capacity <= volumeMO.allocCapacity ? SZERO : CapacityUtil.convert512BToCap(volumeMO
                    .capacity - volumeMO.allocCapacity);

        this.allocType = volumeMO.allocType.toString();
        this.status = volumeMO.status.toString();
        this.capacityUsage = volumeMO.capacity <= volumeMO.allocCapacity ? 1.0 : (volumeMO.allocCapacity * 1.0) /
                volumeMO.capacity;
        this.storagePoolId = volumeMO.storagePoolId;
        return this;
    }

    /**
     * get storage pool mo and fill the information
     *
     * @param storagePoolMO storage Pool mo
     */
    public void updateWithPool(StoragePoolMO storagePoolMO) {
        this.storagePoolName = storagePoolMO.name;
        this.storagePoolId = storagePoolMO.id;
        this.storagePoolCapactiy = CapacityUtil.convert512BToCap(storagePoolMO.totalCapacity);
        this.storagePoolFreeCap = CapacityUtil.convert512BToCap(storagePoolMO.freeCapacity);
        this.storagePoolUsage = 1 - ((storagePoolMO.freeCapacity * 1.0) / storagePoolMO.totalCapacity);
    }

    /**
     * update the device info
     * @param deviceInfo Device info
     */
    public void updateWithStorage(DeviceInfo deviceInfo) {
        this.storageId = deviceInfo.uid;
        this.storageType = deviceInfo.deviceType;
        this.storageIP = deviceInfo.ip;
    }


    /**
     * uptate with iscsi identifer
     *
     * @param scsiLun
     * @return
     */
    public boolean updateWithScsiLun(ScsiLun scsiLun) throws NullPointerException{
        try {
            if (!"disk".equals(scsiLun.getDeviceType())) {
                return false;
            }
            if (scsiLun instanceof HostScsiDisk) {
                this.setLocal(((HostScsiDisk) scsiLun).isLocalDisk());
            }
            this.setIdentifier(scsiLun.getCanonicalName().trim());
            return true;
        } catch (NullPointerException ex) {
            throw ex;
        }
    }
}
