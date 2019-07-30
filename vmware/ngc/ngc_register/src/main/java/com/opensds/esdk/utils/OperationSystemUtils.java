package com.opensds.esdk.utils;

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
