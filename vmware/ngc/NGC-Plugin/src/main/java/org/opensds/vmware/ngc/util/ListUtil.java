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

package org.opensds.vmware.ngc.util;

import java.util.Collections;
import java.util.List;

public class ListUtil {

    /**
     * get the safe sub list of Utils
     * @param list list
     * @param fromIndex page start
     * @param toIndex page count
     * @param <T> list
     * @return list form start to start+toIndex
     */
    public static <T> List<T> safeSubList(List<T> list, int fromIndex, int toIndex) {
        int size = list.size();
        if (fromIndex >= size || toIndex <= 0 || fromIndex >= toIndex) {
            return Collections.emptyList();
        }
        fromIndex = Math.max(0, fromIndex);
        toIndex = Math.min(size, toIndex);
        return list.subList(fromIndex, toIndex);
    }
}
