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

package org.opensds.vmware.ngc.utils;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.io.File;
import java.net.URISyntaxException;
import java.net.URL;

public class OperationSystemUtils {

    private static final int DISKSPACE_THRESHOLD = 629145600;

    private static final Logger _logger = LogManager.getLogger(OperationSystemUtils.class);

    private OperationSystemUtils() {
    }

    /**
     * get free disk space from host
     * @return
     */
    public static boolean isFreeDiskSpace() {
        File file;
        try {

            String path = FileUtils.getProjectPath();
            File rootFile = new File(path);
            long freeSpace = rootFile.getFreeSpace();
            if (freeSpace < DISKSPACE_THRESHOLD) {
                _logger.error("Space is not enough! Free disk space is: " + freeSpace);
                return true;
            }
            return false;
        }
        catch (Exception urle) {
            _logger.error("File to url error", urle);
            return false;
        }
    }
}
