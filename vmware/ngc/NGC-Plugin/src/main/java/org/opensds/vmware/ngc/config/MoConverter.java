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

package org.opensds.vmware.ngc.config;

import com.vmware.vim25.ManagedObjectReference;
import org.springframework.core.convert.converter.Converter;
import java.util.Locale;

public class MoConverter implements Converter<String, ManagedObjectReference> {

    @Override
    public ManagedObjectReference convert(String str) {
        return getMoFromUId(str);
    }

    /**
     * get mob object from UID
     * @param moId string
     * @return ManagedObjectReference
     */
    public static ManagedObjectReference getMoFromUId(String moId) {
        ManagedObjectReference moRef = new ManagedObjectReference();
        String[] moData = moId.split(":");
        if (moData.length < 2) {
            throw new IllegalArgumentException(String.format(Locale.ROOT, "The moId is not valid :{}", moId));
        }
        String moType = moData[0];
        String moValue = moData[1];
        moRef.setType(moType);
        moRef.setValue(moValue);
        return moRef;
    }
}
