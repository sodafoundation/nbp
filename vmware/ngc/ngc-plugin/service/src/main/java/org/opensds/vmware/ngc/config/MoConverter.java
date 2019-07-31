package org.opensds.vmware.ngc.config;

import com.vmware.vim25.ManagedObjectReference;
import org.springframework.core.convert.converter.Converter;

public class MoConverter implements Converter<String, ManagedObjectReference> {
    @Override
    public ManagedObjectReference convert(String s) {
        return getMoFromUId(s);
    }
    public static ManagedObjectReference getMoFromUId(String moId) {
        ManagedObjectReference moRef = new ManagedObjectReference();
        String[] moData = moId.split(":");
        if (moData.length < 2) {
            throw new RuntimeException(String.format("The moId is not valid :{}", moId));
        }
        String moType = moData[0];
        String moValue = moData[1];
        moRef.setType(moType);
        moRef.setValue(moValue);
        return moRef;
    }
}
