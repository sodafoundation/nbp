package org.opensds.vmware.ngc.util;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.List;



public class FilterUtils {

    private static final Log logger = LogFactory.getLog(FilterUtils.class);

    /**
     * filter the list , get the want volumes by values
     * @param list
     * @param type
     * @param value
     * @param <T>
     * @return
     */
    public static <T> List<T> filterList(List<T> list, String type, String value) {

        if (type == null || type.isEmpty()) {
            return list;
        }

        List<T> relist = new ArrayList<>();
        for (T tMo : list) {
            try {
                Field field = tMo.getClass().getDeclaredField(type.toLowerCase());
                field.setAccessible(true);
                Object object = field.get(tMo);

                if (object instanceof String) {
                    String strObject = (String)object;
                    if (strObject.contains(value)) {
                        relist.add(tMo);
                    }
                }
            } catch (NoSuchFieldException | SecurityException ex) {
                logger.error("Filter list error :" + ex.getMessage());
            } catch (IllegalAccessException ex) {
                logger.error("ex :" + ex.getMessage());
            }
        }
        return relist;
    }
}
