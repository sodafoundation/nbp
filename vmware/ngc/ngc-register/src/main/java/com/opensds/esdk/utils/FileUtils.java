package com.opensds.esdk.utils;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import java.io.File;
import java.net.URL;
import java.net.URLDecoder;


public class FileUtils {

    private static final Logger _logger = LogManager.getLogger(FileUtils.class);

    public static String getBasePath() {
        URL url = FileUtils.class.getResource("/");//classes
        if (null == url) {
            _logger.error("FileManager.class.getResource(\"/\") return null.");
            return "";
        }
        File file = new File(url.getFile());
        return file.getParentFile().getParent() + File.separator;
    }

    public static String getParentPath() {
        URL url = FileUtils.class.getResource("/");//classes
        if (null == url) {
            _logger.error("FileManager.class.getResource(\"/\") return null.");
            return "";
        }
        File file = new File(url.getFile());
        String filePath = file.getParentFile().getParentFile().getParent();
        if (filePath != null && filePath.toCharArray()[0] == '\\') {
            String tmp = filePath.substring(1, filePath.length());
            return tmp + File.separator;
        }
        return file.getParentFile().getParentFile().getParent() + File.separator;
    }

    public static String getConfigPath() {
        URL url = FileUtils.class.getResource("/");//classes
        if (null == url) {
            _logger.error("FileManager.class.getResource(\"/\") return null.");
            return "";
        }
        File file = new File(url.getFile());
        return file.getParentFile().getParent() + File.separator + "config";
    }

    public static String getDirectPath() {
        String filePath = System.getProperty("java.class.path");
        String pathSplit = System.getProperty("path.separator");

        if (filePath.contains(pathSplit)){
            filePath = filePath.substring(0, filePath.indexOf(pathSplit));
        } else if (filePath.endsWith(".jar")) {
            filePath = filePath.substring(0, filePath.lastIndexOf(File.separator) + 1);

        }
        return filePath;
    }

    public static String getProjectPath() {
        String availdStr = File.separator + "app";
        return System.getProperty("user.dir").replace(availdStr, "");
    }

    public static String getClassPath() {
        return FileUtils.class.getProtectionDomain().getCodeSource().getLocation().getPath();
    }

    public static String getJarPath() {
        URL url = FileUtils.class.getProtectionDomain().getCodeSource().getLocation();
        String filePath = null;
        try {
            filePath = URLDecoder.decode(url.getPath(), "utf-8");
        } catch (Exception e) {
            e.printStackTrace();
        }
        if (filePath.endsWith(".jar")) {
            filePath = filePath.substring(0, filePath.lastIndexOf("/") + 1);
        }

        File file = new File(filePath);

        // /If this abstract pathname is already absolute, then the pathname
        // string is simply returned as if by the getPath method. If this
        // abstract pathname is the empty abstract pathname then the pathname
        // string of the current user directory, which is named by the system
        // property user.dir, is returned.
        filePath = file.getAbsolutePath();
        return filePath;
    }
}
