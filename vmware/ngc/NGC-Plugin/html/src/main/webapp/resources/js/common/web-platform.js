// ------------------------------------------------------------------------------
// Javascript initialization to include when using the HTML bridge:
// - Creates the plugin's private namespace org_opensds_storage_devices
// - Defines APIs to ensure compatibility with future Web Client HTML platform
// ------------------------------------------------------------------------------

// WEB_PLATFORM is the VMware Web Client platform reference.
// When the Flex client is running it is defined as the Flash container.
var WEB_PLATFORM = self.parent.WEB_PLATFORM;
var isChromeBrowser = (window.navigator.userAgent.indexOf("Chrome/") >= 0);
var isFlexClient = !!self.parent.document.getElementById("container_app");

if (!WEB_PLATFORM || (isChromeBrowser && isFlexClient)) {
    WEB_PLATFORM = self.parent.document.getElementById("container_app");
    if (isChromeBrowser) {
        WEB_PLATFORM = Object.create(WEB_PLATFORM);
    }

    self.parent.WEB_PLATFORM = WEB_PLATFORM;

    // The web context starts with a different root path depending on which client is running.

    if (!WEB_PLATFORM.getRootPath) {
        WEB_PLATFORM.getRootPath = function () {
           return "/vsphere-client";
        }
    }
    // Declare unknown client type explicitly.
    if (!WEB_PLATFORM.getClientType) {
        WEB_PLATFORM.getClientType = function () {
           return "flex";
        }
    }
    // Declare unknown client version explicitly.
    if (!WEB_PLATFORM.getClientVersion) {
        WEB_PLATFORM.getClientVersion =  function () {
           return "6.0";
        }
    }
}

var getClientType = WEB_PLATFORM.getClientType;

// Define a private namespace using the plugin bundle name,
// It should be the only global symbol added by this plugin!
var org_opensds_storage_devices;
if (!org_opensds_storage_devices) {
    org_opensds_storage_devices = {};

    // The web context path to use for server requests
    // (same as the Web-ContextPath value in the plugin's MANIFEST.MF)

    //fix: IN IE 11 , getClientType has the emerge error "不能执行已释放 Script 的代码";
    try{
        org_opensds_storage_devices.webContextPath = WEB_PLATFORM.getRootPath() + "/opensds";
    }
    catch (err){
        org_opensds_storage_devices.webContextPath = "/vsphere-client/opensds";
        console.log("getRootPath error: " + err);
    }


    //fix: IN IE 11 , getClientType has the emerge error "不能执行已释放 Script 的代码";
    if (!WEB_PLATFORM){
        WEB_PLATFORM = self.top.document.getElementById("container_app");
    }
    try {
        if(!getClientType){
           getClientType = getClientTypeFun;
        }
        if(getClientType() == "flex") {
            org_opensds_storage_devices.baseURL = "/vsphere-client/";
        }else{
            org_opensds_storage_devices.baseURL = "/ui/";
        }
    }catch (err) {
       org_opensds_storage_devices.baseURL = "/vsphere-client/";
       console.log("getClientType error: " + err);
     }

    org_opensds_storage_devices.deviceNS = "org.opensds.vmware.ngc.device";

    // The API setup is done inside an anonymous function to keep things clean.
    // See the HTML bridge documentation for more info on those APIs.
    (function () {
        // Namespace shortcut
        var ns = org_opensds_storage_devices;

        // ------------------------ Private functions -------------------------------

        // Get a string from the resource bundle defined in plugin.xml
        function getString(key, params) {
            var result =  WEB_PLATFORM.getString("org_opensds_storage_devices", key, params);
            if(result == null){
                return key;
            }else{
                return result;
            }
        }

        // Get a parameter value from the current document URL
        function getURLParameter(name) {
            return (new RegExp('[?|&]' + name + '=' + '([^&;]+?)(&|#|;|$)')
                .exec(location.href) || [, ""])[1].replace(/\+/g, '%20') || null;
        }

        //  function getURLParameter(name) {
        //      var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)");
        //      var r = window.location.search.substr(1).match(reg);
        //      if(r!=null)return  unescape(r[2]); return null;
        //  }

        // Build the REST url prefix to retrieve a list of properties,
        // this is mapped to the DataAccessController on the java side.
        function buildDataUrl(objectId, propList) {
            var propStr = propList.toString();
            var dataUrl = ns.webContextPath +
                "/rest/data/properties/" + objectId + "?properties=" + propStr;
            return dataUrl;
        }

        function buildDataSystemUrl(objectId) {
            var dataUrl = ns.webContextPath +
                "/rest/data/system/" + objectId;
            return dataUrl;
        }

        function buildDataStoragePoolUrl(objectId) {
            var dataUrl = ns.webContextPath +
                "/rest/data/storagepool/" + objectId;
            return dataUrl;
        }

        function buildDataStorageLunUrl(objectId, propList) {
            var propStr = propList.toString();
            var dataUrl = ns.webContextPath +
                "/rest/data/storagelun/" + objectId + "?storagepool=" + propStr;
            return dataUrl;
        }

        function buildDataStorageAlarmsUrl(objectId) {
            var dataUrl = ns.webContextPath +
                "/rest/data/storagealarms/" + objectId;
            return dataUrl;
        }

        // -------------------------- Public APIs --------------------------------

        // Functions exported to the org_opensds_storage_devices namespace
        ns.getString = getString;
        ns.buildDataUrl = buildDataUrl;
        ns.buildDataSystemUrl = buildDataSystemUrl;
        ns.buildDataStoragePoolUrl = buildDataStoragePoolUrl;
        ns.buildDataStorageLunUrl = buildDataStorageLunUrl;
        ns.buildDataStorageAlarmsUrl = buildDataStorageAlarmsUrl;

        // APIs added to WEB_PLAFORM for compatibility with future HTML platform
        if (!WEB_PLATFORM) {
             WEB_PLATFORM = self.top.document.getElementById("container_app");
        }

        // Get the current context object id or return null if none is defined
        WEB_PLATFORM.getObjectId = function () {
            return getURLParameter("objectId");
        };
        // Get the current action Uid or return null if none is defined
        WEB_PLATFORM.getActionUid = function () {
            return getURLParameter("actionUid");
        };
        // Get the comma-separated list of object ids for an action, or null for a global action
        WEB_PLATFORM.getActionTargets = function () {
            return getURLParameter("targets");
        };

        WEB_PLATFORM.getHostId = function () {
                    return getURLParameter("hostId");
        };

        WEB_PLATFORM.getServerGuid= function () {
                    return getURLParameter("serverGuid");
        };
        // Get the current locale
        WEB_PLATFORM.getLocale = function () {
            return getURLParameter("locale");
        };

        // Get the info provided in a global view using a vCenter selector
        WEB_PLATFORM.getVcSelectorInfo = function () {
            var info = {
                serviceGuid: getURLParameter("serviceGuid"),
                serverGuid: getURLParameter("serverGuid"),
                serviceUrl: getURLParameter("serviceUrl")
            };
            return info;
        };

        // Set a refresh handler called when the user hits Refresh in the WebClient top toolbar.
        // This is the Flex Client implementation which doesn't need document as 2nd argument.
        if (isIE()){
            WEB_PLATFORM.setGlobalRefreshHandler = function (handler) {
                WEB_PLATFORM["refresh" + window.name] = handler;
            };
        } else {
            if (WEB_PLATFORM.getClientType() === "flex") {
                WEB_PLATFORM.setGlobalRefreshHandler = function (handler) {
                    WEB_PLATFORM["refresh" + window.name] = handler;
                };
            }
        }
    })();
} // end of if (!org_opensds_storage_devices)


function isIE() { //ie?
    if (!!window.ActiveXObject || "ActiveXObject" in window)
        return true;
    else
        return false;
}
