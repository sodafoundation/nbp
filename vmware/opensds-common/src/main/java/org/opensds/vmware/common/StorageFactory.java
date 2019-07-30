package org.opensds.vmware.common;

import org.opensds.vmware.common.adapters.thirdparty.oceanstor.OceanStor;

import java.lang.reflect.Constructor;

public class StorageFactory {
    public static String[] listStorages() {
        return new String[]{
            OceanStor.class.getName(),
        };
    }

    public static Storage newStorage(String type, String name) throws Exception {
        Class cls = Class.forName(type);
        Constructor constructor = cls.getDeclaredConstructor(String.class);
        return (Storage) constructor.newInstance(name);
    }
}
