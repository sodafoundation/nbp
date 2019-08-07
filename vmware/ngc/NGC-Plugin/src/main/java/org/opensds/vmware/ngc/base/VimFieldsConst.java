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

package org.opensds.vmware.ngc.base;

public interface VimFieldsConst {

    interface MoTypesConst {
        String Alarm = "Alarm";
        String AlarmManager = "AlarmManager";
        String AuthorizationManager = "AuthorizationManager";
        String ClusterComputeResource = "ClusterComputeResource";
        String ComputeResource = "ComputeResource";
        String CustomFieldsManager = "CustomFieldsManager";
        String CustomizationSpecManage = "CustomizationSpecManager";
        String Datacenter = "Datacenter";
        String Datastore = "DatastoreInfo";
        String DiagnosticManager = "DiagnosticManager";
        String EnvironmentBrowser = "EnvironmentBrowser";
        String EventHistoryCollector = "EventHistoryCollector";
        String EventManager = "EventManager";
        String Folder = "Folder";
        String HistoryCollector = "HistoryCollector";
        String HostAutoStartManager = "HostAutoStartManager";
        String HostCpuSchedulerSystem = "HostCpuSchedulerSystem";
        String HostDatastoreBrowser = "HostDatastoreBrowser";
        String HostDatastoreSystem = "HostDatastoreSystem";
        String HostDiagnosticSystem = "HostDiagnosticSystem";
        String HostDiskManagerLease = "HostDiskManagerLease";
        String HostFirewallSystem = "HostFirewallSystem";
        String HostLocalAccountManager = "HostLocalAccountManager";
        String HostMemorySystem = "HostMemorySystem";
        String HostNetworkSystem = "HostNetworkSystem";
        String HostServiceSystem = "HostServiceSystem";
        String HostSnmpSystem = "HostSnmpSystem";
        String HostStorageSystem = "HostStorageSystem";
        String HostSystem = "HostSystem";
        String HostVMotionSystem = "HostVMotionSystem";
        String LicenseManager = "LicenseManager";
        String ManagedEntity = "ManagedEntity";
        String Network = "Network";
        String OptionManager = "OptionManager";
        String PerformanceManager = "PerformanceManager";
        String PropertyCollector = "PropertyCollector";
        String PropertyFilter = "PropertyFilter";
        String ResourcePool = "ResourcePool";
        String ScheduledTask = "ScheduledTask";
        String ScheduledTaskManager = "ScheduledTaskManager";
        String SearchIndex = "SearchIndex";
        String ServiceInstance = "ServiceInstance";
        String SessionManager = "SessionManager";
        String Task = "Task";
    }
    interface PropertyNameConst {
        interface HostSystem {
            String Datastore = "datastore";
            String Config = "config";
            String VM = "vm";
            String Runtime = "runtime";
            String ConfigManager = "configManager";
            String HardWare = "hardware";
            String Name = "name";
        }

        interface Datastore {
            String Info = "info";
            String Summary = "summary";
            String OverallStatus = "overallStatus";
            String Host = "host";
            String Name = "name";
        }

        interface VM{
            String Config = "config";
            String Runtime = "runtime";
        }
    }

}
