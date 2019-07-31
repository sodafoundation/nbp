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

package org.opensds.vmware.ngc.expections;

import org.opensds.vmware.ngc.entity.ResultInfo;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

public class ExpectionHandle {

    private static Log _logger = LogFactory.getLog(ExpectionHandle.class);

    public static void handleExceptions(ResultInfo resultInfo, Throwable e) {
        _logger.error(e.getMessage());
        resultInfo.setMsg(e.getMessage());
        if (e instanceof InactiveSessionException){
            resultInfo.setStatus("error");
        }
    }
}
