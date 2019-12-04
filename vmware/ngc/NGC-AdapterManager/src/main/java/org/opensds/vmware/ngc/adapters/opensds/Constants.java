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

package org.opensds.vmware.ngc.adapters.opensds;

import java.util.Properties;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

 enum UNIT_TYPE {

    /**
     * byte
     */
    Byte(1L),

    /**
     * KB
     */
    KB(1024L),

    /**
     * MB
     */
    MB(1024L * 1024L),

    /**
     * GB
     */
    GB(1024L * 1024L * 1024L),

    /**
     * TB
     */
    TB(1024L * 1024L * 1024L * 1024L),

    /**
     * PB
     */
    PB(1024L * 1024L * 1024L * 1024L * 1024L);

    private long scale;

    UNIT_TYPE(long scale)
    {
        this.scale = scale;
    }


    public long getUnit()
    {
        return this.scale;
    }
}

public enum Constants {
	OPENSDS_TENANT,
	OPENSDS_DOMAIN,
	OPENSDS_AVAILABILITYZONE,
	OPENSDS_STORAGENAME,
	OPENSDS_VENDOR,
	OPENSDS_HOST_ACCESSMODE;

	private static final String PATH = "/constants.properties";

	private static Properties properties;

	private String value;

    private static final Log logger = LogFactory.getLog(Constants.class);

	private void init() {
		if (properties == null) {
			properties = new Properties();
			try {
				properties.load(this.getClass().getResourceAsStream(PATH));
			}
			catch (Exception ex) {
				logger.error(String.format("Error in loading Constant Properties, Error Message is: %s", ex));
			}
		}
		value = (String) properties.get(this.toString());
	}

	public String getValue() {
		if (value == null) {
			init();
		}
		return value;
	}
}

enum VOLUME_STATUS {
	AVAILABLE("available"), INUSE("inUse"), ERROR("error");

	private String value;

	private VOLUME_STATUS(String v) {
		this.value = v;
	}

	public String getValue() {
		return this.value;
	}
}
