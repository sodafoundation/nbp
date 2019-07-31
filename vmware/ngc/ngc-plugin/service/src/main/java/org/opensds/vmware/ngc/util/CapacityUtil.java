package org.opensds.vmware.ngc.util;

import org.opensds.vmware.ngc.base.CapabilityUnitTypeEnum;

import java.math.RoundingMode;
import java.text.DecimalFormat;

public final class CapacityUtil {
    private static final Long UNITS = 1024L;

    private static long NUM_1024 = 1024;

    private static long TB = NUM_1024 * NUM_1024 * NUM_1024 * NUM_1024;

    private static long GB = NUM_1024 * NUM_1024 * NUM_1024;

    private static long MB = NUM_1024 * NUM_1024;

    private static long KB = NUM_1024;

    private CapacityUtil() {
    }

    public static String convert512BToCap(long capacity) {
        DecimalFormat decimalFormat = new DecimalFormat(".000");
        decimalFormat.setRoundingMode(RoundingMode.FLOOR);
        if (0 == capacity) {
            return "0.000" + " MB";
        }
        if (capacity / TB >= 1) {
            return (decimalFormat.format(capacity / (double) TB)).toString() + " TB";
        }
        else if (capacity / GB >= 1) {
            return (decimalFormat.format(capacity / (double) GB)).toString() + " GB";
        }
        else if (capacity / MB >= 1) {
            return (decimalFormat.format(capacity / (double) MB)).toString() + " MB";
        }
        else if (capacity / KB >= 1) {
            return (decimalFormat.format(capacity / (double) KB)).toString() + " KB";
        }
        else {
            return (decimalFormat.format(capacity)).toString() + " B";
        }
    }

    public static String converByteToGiga(Long capacity) {
        Long units = UNITS * UNITS * UNITS;
        return String.valueOf(capacity / units);
    }

    public static long converGBToByte(long capacity) {
        Long units = UNITS * UNITS * UNITS;
        return capacity*units;
    }

    public static String convertByteToCap(Long capacity) {

        if (0 >= capacity) {
            return "0.000" + " " + CapabilityUnitTypeEnum.MB.toString();
        }
        int scope = MathUtil.get2M(capacity);
        CapabilityUnitTypeEnum capUnit = CapabilityUnitTypeEnum.getCapabilityUnitTypeByorder(scope
                / 10);
        double tmp = (double) capacity / capUnit.getUnit();
        String reTurnCap = MathUtil.downScaleToString(tmp, 3);

        return reTurnCap + " " + capUnit.toString();
    }

    public static long _calCapRate (Long itemCap, Long totalCap) {
        if (0 == itemCap || 0 == totalCap) {
            return 0;
        }
        Long result = itemCap * 100 / totalCap;
        if (0 == result) {
            return 1;
        }

        return result;
    }

    public static Long convertCapToLong (String capacity){
        if(null == capacity || "".equals(capacity)){
            return 0L;
        }
        double cap = 0.0;
        if(capacity.indexOf("TB") != -1){
            cap = Double.parseDouble(capacity.replace("TB", ""))*TB;
        } else if(capacity.indexOf("GB") != -1){
            cap = Double.parseDouble(capacity.replace("GB", ""))*GB;
        } else if(capacity.indexOf("MB") != -1){
            cap = Double.parseDouble(capacity.replace("MB", ""))*MB;
        }else if(capacity.indexOf("KB") != -1){
            cap = Double.parseDouble(capacity.replace("KB", ""))*KB;
        } else{
            return 0L;
        }
        return new Double(cap).longValue();
    }
}
