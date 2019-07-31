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
