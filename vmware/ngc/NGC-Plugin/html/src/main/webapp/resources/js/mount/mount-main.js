function initSelect() {
    var url = ns.webContextPath + "/rest/device/getList?t=" + new Date();
    $.ajax({
        contentType: "application/x-www-form-urlencoded;charset=UTF-8",
        url: encodeURI(url),
        async: false,
        data: {},
        success: function (resp) {
            var arr = eval(resp.data);
            var objSelectNow = document.getElementById("StorageDevice");
            for (var i = 0; i < arr.length; i++) {
                var jsonObj = arr[i];
                var objOption = document.createElement("OPTION");
                objOption.text = jsonObj.ip;
                objOption.value = jsonObj.sn;
                objOption.id = jsonObj.uid;
                if (i == 0) {
                    deviceId = jsonObj.uid;
                }
                objSelectNow.options.add(objOption);
            }
            $("#StorageDevice").unbind();
            $("#StorageDevice").bind("change", function () {
                deviceId = $('#StorageDevice').find("option:selected").attr("id");
                refreshData();
            });
        },
        complete: function (XHR, TS) {
            XHR = null;
        }
    });
}


function refreshData() {
    loadpage2_data_params = "&deviceId=" + deviceId + "&filterType=" + filterType
        + "&filterValue=" + filterValue + "&serverGuid=" + serverGuid + "&t="
        + new Date();
    $("#chk_all").prop("checked", false);
    $("#chk_all").attr("disabled", "disabled");
    $("#btnMount").prop("disabled", "disabled");
    $("#diverrorLUN").hide();
    if ($("#" + divhead_id).length > 0) {
        $("#" + divhead_id).width($("#divMain").width() - 22);
    }
    $('#unmappedlunTabFrame').prop("src", "");
    $("#pager1").remove();
    if (deviceId == "") {
        return;
    }
    $("#divLoadingMappedLUN").css("display", "block");
    var url = ns.webContextPath + "/rest/data/host/mountableVolumeList/count/" + hostId
        + "?deviceId=" + deviceId + "&filterType=" + filterType
        + "&filterValue=" + filterValue + "&serverGuid=" + serverGuid + "&t="
        + new Date();
    $("#mappedLunList").bigPage(
        {
            container: "pager1",
            ajaxData: {
                url: encodeURI(url),
                params: {
                    loaddingId: "divLoadingMappedLUN",
                    errorloaddingId: "diverrorLUN",
                    iframeId: "unmappedlunTabFrame",
                    data_url: ns.webContextPath + "/resources/html/mount/lunTab.html",
                    data_params: loadpage2_data_params
                }
            },
            pageSize: pagesize_lun,
            toPage: toPage_lun,
            position: "down",
            callback: null
        });
}


function changesize() {
    var divMainHeight = $("#divMain").height();
    var topHeight = $("#top").height();
    var lineTop = 5 + topHeight + 5;
    var buttonsTop = lineTop + 2 + 5;
    $("#line").css("top", lineTop);
    $("#buttons").css("top", buttonsTop);
    var buttonsHeight = $("#buttons").height();
    var tableTop = buttonsTop + buttonsHeight + 5;
    var tableHeight = divMainHeight - tableTop - 3;
    $("#mappedLunListDiv").height(tableHeight - 100);

    $("#unmappedlunTabFrame").height($("#mappedLunListDiv").height() - $("#mappedLunList").height() - 24);
}
