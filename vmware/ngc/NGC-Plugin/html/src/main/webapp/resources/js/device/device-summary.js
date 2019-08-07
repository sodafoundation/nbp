// Use JQuery's $(document).ready to execute the script when the document is loaded.
// All variables and functions are also hidden from the global scope.
function getParam(name) {
    return (new RegExp('[?|&]' + name + '=' + '([^&;]+?)(&|#|;|$)')
            .exec(location.href) || [,""])[1].replace(/\+/g, '%20') || null;
}


$(document).ready(
    function () {
        makeHelp();
        //changesize();
        // Namespace shortcut
        var ns = org_opensds_storage_devices;

        // Get current object and return if document is loaded before
        // context is set
        var deviceId = getParam("objectId");
        if (!deviceId) {
            return;
        }
        // Data url to get System Info
        //var dataUrl = ns.buildDataSystemUrl(deviceId);
        // dataUrl = dataUrl + "?t=" + new Date();
        var dataUrl = ns.baseURL + "opensds/rest/device/get?deviceID=" + deviceId;

        // Do the actual call now and save as GlobalRefresh handler
        refreshData();
        // The view refreshData function calls to the DataAccessController and returns Json data to insert in the document.
        function refreshData() {
            $("#systemInfoTable").hide();
            $.getJSON(encodeURI(dataUrl), function (resp) {

                if (resp.status=="error") {
                    var description = "<span style='width: 0; height: 100%; display: inline-block; vertical-align: middle;'></span>" + resp.msg;
                    $("#diverrorLun").html(description);
                    return;
                }

                $("#systemInfoTable").show();
//                var jsonObj = eval(data);
                var data = resp.data;

                $("#sys-sn").text(data.sn);
                $("#sys-sn")[0].title = data.sn;

                $("#sys-model").text(data.deviceModel);
                $("#sys-model")[0].title = data.deviceModel;

                $("#sys-ip").text(data.ip);
                $("#sys-ip")[0].title = data.ip;

                $("#sys-status").text(data.deviceStatus);
                $("#sys-status")[0].title = data.deviceStatus;
                if (data.deviceStatus == "normal") {
                    $("#status-img").prop("src", "../../../assets/images/normalstate.png");
                } else {
                    $("#status-img").prop("src", "../../../assets/images/offline.png");
                }
                $("#status-img").show();

                $("#sys-location").text(data.location);
                $("#sys-location")[0].title = data.location;

            });
        }


    });

