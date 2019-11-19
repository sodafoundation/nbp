# OpenSDS vRealize Orchestrator Plugin 

**********************************************************************************
OpenSDS vRealize Orchestrator Plugin 
**********************************************************************************

I. General Information 

    Name:     OpenSDS_Storage_vRO_Plugin_V2.1.10
    Category: vRealize Orchestrator Plugin
    version : 2.1.10
    
II. Description

    The OpenSDS_Storage_vRO Plugin , complying with VMware vRealize Orchestrator API standards, is a workflow 
    automation plug-in for managing Storage devices.

III. Supported Software Version
    
    
    VMware vRealize Orchestrator：7.3~7.5
    VMware vSphere：6.0\6.5\6.7

    
IV.Software Requirements
    
    JRE：1.8
    Mavent: 3.5.4
    Ant: 1.9.10
    
V. Supported Device

    All Southbound Stroage devices supported by OpenSDS.
    
VI. Notice
    
    Add in o11n-util-6.x.x.jar, o11n-security-6.x.x.jar and o11n-sdkapi-6.x.x.jar to the lib folder. these jars can be obtained from VRO appliance /usr/lib/vco/downloads/vco-repo/com/vmware/o11n/o11n-sdkapi/
VII. Build

   To build the plugin package run /run/run.bat file in a windows machine.