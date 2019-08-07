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

import java.io.File;
import java.net.URL;

public class FileManager
{
	public final static String DEVICE_PARENTDIR_NAME = "server";

    public static String getBasePath()
    {
        URL url = FileManager.class.getResource("/");// classes
        File file = new File(url.getFile());
        return file.getPath() + File.separator;
    }

    public static String getDeviceConfigPath()
    {
        URL url = FileManager.class.getResource("/");// classes
        File file = new File(url.getFile());
        return file.getParentFile().getParentFile().getParentFile().getParentFile().
                getParentFile().getParentFile().getParentFile().getPath() + File.separator;
    }

    public static String getDeviceConfigLibPath()
    {
        URL url = FileManager.class.getResource("/");// classes
        File file = new File(url.getFile());
        return file.getParentFile().getParentFile().getParentFile().getParentFile().
                getParentFile().getParentFile().getParentFile().getParentFile().
                getParentFile().getPath() + File.separator;
    }

}
