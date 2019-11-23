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

package org.opensds.vmware.ngc.adapter;

import org.opensds.vmware.ngc.dao.DeviceRepository;
import org.opensds.vmware.ngc.model.DeviceInfo;
import org.opensds.vmware.ngc.base.Constant;
import com.vmware.vise.data.Constraint;
import com.vmware.vise.data.PropertySpec;
import com.vmware.vise.data.ResourceSpec;
import com.vmware.vise.data.query.Comparator;
import com.vmware.vise.data.query.type;
import com.vmware.vise.data.query.DataProviderAdapter;
import com.vmware.vise.data.query.Response;
import com.vmware.vise.data.query.RequestSpec;
import com.vmware.vise.data.query.QuerySpec;
import com.vmware.vise.data.query.ResultSet;
import com.vmware.vise.data.query.ResultItem;
import com.vmware.vise.data.query.ObjectIdentityConstraint;
import com.vmware.vise.vim.data.VimObjectReferenceService;
import com.vmware.vise.data.query.CompositeConstraint;
import com.vmware.vise.data.query.PropertyConstraint;
import com.vmware.vise.data.query.PropertyValue;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import java.util.Arrays;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.HashSet;
import java.util.Iterator;
import java.util.Map;
import java.util.Locale;

@type(Constant.DEVICE_TYPE)
@Component
public class DeviceDataAdapter implements DataProviderAdapter {

    /**
     * device type
     */
    public static final String DEVICE_TYPE = Constant.DEVICE_TYPE;

    private static final Log logger = LogFactory.getLog(DeviceDataAdapter.class);

    @Autowired
    private VimObjectReferenceService objectRefService;

    @Autowired
    private DeviceRepository deviceRepository;

    /**
     * get data from RequestSpec
     * @param request RequestSpec
     * @return Response
     */
    @Override
    public Response getData(RequestSpec request) {
        if (request == null) {
            throw new IllegalArgumentException("request must be non-null.");
        }
        QuerySpec[] querySpecs = request.querySpec;
        List<ResultSet> results = new ArrayList<ResultSet>(querySpecs.length);
        for (QuerySpec qs : querySpecs) {
            ResultSet rs = processQuery(qs);
            results.add(rs);
        }

        Response response = new Response();
        response.resultSet = results.toArray(new ResultSet[]{});
        return response;
    }

    private ResultSet processQuery(QuerySpec qs) {
        ResultSet rs = new ResultSet();
        if (!validateQuerySpec(qs)) {
            return rs;
        }
        List<ResultItem> items = processConstraint(qs.resourceSpec.constraint, qs.resourceSpec.propertySpecs);
        if (null != items) {
            rs.totalMatchedObjectCount = items.size();
        } else {
            rs.totalMatchedObjectCount = 0;
        }
        ResultItem[] rt = new ResultItem[]{};
        if (null != rt && null != items) {
            rs.items = items.toArray(rt);
        }

        rs.queryName = qs.name;
        return rs;
    }

    private String[] convertPropertySpec(PropertySpec[] pSpecs) {
        Set<String> properties = new HashSet<String>();
        if (pSpecs != null) {
            for (PropertySpec pSpec : pSpecs) {
                for (String property : pSpec.propertyNames) {
                    properties.add(property);
                }
            }
        }
        return properties.toArray(new String[]{});
    }

    private List<ResultItem> processConstraint(Constraint constraint, PropertySpec[] propertySpecs) {
        List<ResultItem> items = null;

        if (constraint instanceof ObjectIdentityConstraint) {
            ObjectIdentityConstraint oic = (ObjectIdentityConstraint) constraint;
            items = processObjectIdentityConstraint(oic, propertySpecs);
        } else if (constraint instanceof CompositeConstraint) {
            CompositeConstraint cc = (CompositeConstraint) constraint;
            items = processCompositeConstraint(cc, propertySpecs);
        } else if (constraint instanceof PropertyConstraint) {
            PropertyConstraint pc = (PropertyConstraint) constraint;
            items = processPropertyConstraint(pc, propertySpecs);
        } else if (isSimpleConstraint(constraint)) {
            items = processSimpleConstraint(constraint, propertySpecs);
        } else {
            logger.error("can not find any property process.");
        }
        return items;
    }

    private List<ResultItem> processCompositeConstraint(CompositeConstraint cc, PropertySpec[] propertySpecs) {
        List<ResultItem> items = new ArrayList<ResultItem>();
        for (Constraint constraint : cc.nestedConstraints) {
            List<ResultItem> individualItems = processConstraint(constraint, propertySpecs);
            items.addAll(individualItems);
        }
        return items;
    }

    private List<ResultItem> processSimpleConstraint(Constraint constraint, PropertySpec[] propertySpecs) {
        String[] requestedProperties = convertPropertySpec(propertySpecs);
        logger.debug(String.format(Locale.ROOT, "ProcessSimpleConstraint: %s  %s",
                constraint.targetType, requestedProperties));
        Map<String, DeviceInfo> allDevices = deviceRepository.getAll();
        return processAll(allDevices, requestedProperties);
    }

    private List<ResultItem> processAll(Map<String, DeviceInfo> allDevice, String[] requestedProperties) {
        List<ResultItem> items = new ArrayList<ResultItem>();
        for (Map.Entry<String, DeviceInfo> entry : allDevice.entrySet()) {
            String uid = entry.getKey();
            DeviceInfo deviceInfo = entry.getValue();
            logger.info(String.format(Locale.ROOT, "Trying to get data for device uid{}", uid));
            ResultItem ri = addDeviceResultItem(uid, deviceInfo, requestedProperties);
            if (ri != null) {
                items.add(ri);
            }
        }
        return items;
    }

    private List<ResultItem> processObjectIdentityConstraint(
            ObjectIdentityConstraint constraint, PropertySpec[] propertySpecs) {
        List<ResultItem> items = new ArrayList<ResultItem>();
        String[] requestedProperties = convertPropertySpec(propertySpecs);

        String uid = objectRefService.getUid(constraint.target);
        DeviceInfo deviceInfo = deviceRepository.get(uid);

        if (deviceInfo != null) {
            ResultItem ri = addDeviceResultItem(uid, deviceInfo, requestedProperties);
            if (ri != null) {
                items.add(ri);
            }
        }
        return items;
    }

    private List<ResultItem> processPropertyConstraint(PropertyConstraint pc, PropertySpec[] propertySpecs) {
        assert (pc.comparator == Comparator.EQUALS);
        String comparableValue = pc.comparableValue.toString();
        List<ResultItem> items = new ArrayList<ResultItem>();
        String[] requestedProperties = convertPropertySpec(propertySpecs);
        Map<String, DeviceInfo> currentObjects = deviceRepository.getAll();
        Iterator<String> it = currentObjects.keySet().iterator();

        while (it.hasNext()) {
            String uid = it.next();
            DeviceInfo deviceInfo = currentObjects.get(uid);

            if (comparableValue.equals(deviceInfo.getProperty(pc.propertyName))) {
                ResultItem ri = addDeviceResultItem(uid, deviceInfo, requestedProperties);
                if (ri != null) {
                    items.add(ri);
                }
            }
        }
        return items;
    }

    private ResultItem addDeviceResultItem(String uid, DeviceInfo deviceInfo, String[] requestedProperties) {
        logger.info(String.format(Locale.ROOT, "requestedProperties for device %s are %s",
                deviceInfo.ip, Arrays.toString(requestedProperties)));
        ResultItem ri = new ResultItem();
        Object deviceRef = deviceInfo.getDeviceReference();
        ri.resourceObject = deviceRef;

        List<PropertyValue> propValArr = new ArrayList<PropertyValue>(requestedProperties.length);
        for (int i = 0; i < requestedProperties.length; ++i) {
            String requestedProperty = requestedProperties[i];
            logger.info(String.format(Locale.ROOT, "Processing property:%s", requestedProperty));
            Object value = deviceInfo.getProperty(requestedProperty);
            if (value != null) {
                PropertyValue pv = new PropertyValue();
                pv.resourceObject = deviceRef;
                pv.propertyName = requestedProperty;
                pv.value = value;
                propValArr.add(pv);
            }
        }
        ri.properties = propValArr.toArray(new PropertyValue[propValArr.size()]);
        return ri;
    }

    private Boolean validateQuerySpec(QuerySpec qs) {
        if (qs == null) {
            return false;
        }

        ResourceSpec resourceSpec = qs.resourceSpec;
        if (resourceSpec == null || resourceSpec.constraint == null) {
            return false;
        }
        return validateConstraint(resourceSpec.constraint);
    }

    private Boolean validateConstraint(Constraint constraint) {
        if (constraint instanceof ObjectIdentityConstraint) {
            Object source = ((ObjectIdentityConstraint) constraint).target;
            return (source != null && DEVICE_TYPE.equals(objectRefService.getResourceObjectType(source)));

        } else if (constraint instanceof CompositeConstraint) {
            CompositeConstraint cc = (CompositeConstraint) constraint;
            for (Constraint c : cc.nestedConstraints) {
                if (!validateConstraint(c)) {
                    return false;
                }
            }
            return true;

        } else if (constraint instanceof PropertyConstraint) {
            return DEVICE_TYPE.equals(constraint.targetType)
                    && ((PropertyConstraint) constraint).comparator == Comparator.EQUALS;

        } else if (isSimpleConstraint(constraint)) {
            return (DEVICE_TYPE.equals(constraint.targetType));

        } else {
            return false;
        }
    }

    private Boolean isSimpleConstraint(Object constraint) {
        return (constraint.getClass().getSimpleName().equals(Constraint.class.getSimpleName()));
    }
}
