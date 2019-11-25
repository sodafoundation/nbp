$(document).ready(function () {
    bindEvent();
    initData();
});

function initData() {
    var request = new Object();
    request = GetRequest();
    parent.$("#divLoadingLun").show();
    var url = "";
    if (parent.lun_fs_flag == "lun") {
        url = parent.urlForLun + parent.hostId + "?" +
            "serverGuid=" + parent.serverGuid +
            "&filterType=" + parent.filterType +
            "&filterValue=" + parent.filterValue +
            "&start=" + request["start"] + "&count=" + request["pagesize"] + "&t=" + new Date();
    } else if (parent.lun_fs_flag == "fs") {
        url = parent.urlForFs + parent.hostId + "?" +
            "serverGuid=" + parent.serverGuid +
            "&filterType=" + parent.filterType +
            "&filterValue=" + parent.filterValue +
            "&start=" + request["start"] + "&count=" + request["pagesize"] + "&t=" + new Date();
    }

    var lunReq = new req(url, "");
    var lunhandler = new handler(function doSuccess(resp) {
        if (resp.errorCode) {
            parent.$("#divLoadingLun").hide();
            $("#divError").text(resp.errorDesc).show();
            return;
        }
        a2t("#hostLunTbody", "#cloneLun", resp.data);
        scroll("hostLunTab", "lunTabDiv", 1, parent.divhead_id_lun, "hostLunTable");
        parent.$("#divLoadingLun").hide();
        parent.$('#snapshotFrame').prop("src", "");
        parent.$("#pager2").remove();
        parent.$("#recoverBtn").addClass("disabled");
        parent.$("#recoverBtn").prop("disabled", "disabled");
        parent.$("#refreshSnapBtn").addClass("disabled");
        parent.$("#delSnapBtn").addClass("disabled");
        parent.$("#delSnapBtn").prop("disabled", "disabled");
    }, function doFailed() {
        parent.$("#divLoadingLun").hide();
    });
    sendMsg(lunReq, lunhandler);
}
/**
 * Only provide radio buttons, do not provide bulk snapshot / backup function, if provided and no longer pass the checkbox
 */
function bindEvent() {
    $("#hostLunTable tbody tr").bind("click", function (event) {

        $("#hostLunTable tbody tr td").css("background-color", "#FFFFFF");
        $(this).find('td').each(function (i) {
            $(this).css("background-color", "#abcefc");
        });

        if (parent.lun_fs_flag == "lun") {
            parent.lunObj.id = $(this).find("[name='id']").text();
            parent.lunObj.name = $(this).find("[name='name']").text();
            parent.lunObj.status = $(this).find("[name='status']").text();
            parent.lunObj.usedType = $(this).find("[name='usedType']").text();
        } else if (parent.lun_fs_flag == "fs") {
            parent.fsObj.id = $(this).find("[name='id']").text();
            parent.fsObj.name = $(this).find("[name='name']").text();
            parent.fsObj.usedByStatus = $(this).find("[name='usedByStatus']").text();
            parent.fsObj.datastoreId = $(this).find("[name='datastoreId']").text();
            //for mount
            parent.fsObj.localPath = $(this).find("[name='localPath']").text();
            parent.fsObj.remoteHost = $(this).find("[name='remoteHost']").text();
            parent.fsObj.remotePath = $(this).find("[name='remotePath']").text();
        }

        parent.devObj.id = $(this).find("[name='storageId']").text();

        parent.$("#refreshSnapBtn").prop("disabled", "");
        parent.$("#refreshSnapBtn").removeClass("disabled");
        parent.$("#refreshLunBtn").prop("disabled", "");

        parent.$("#showBackupBtn").prop("disabled", "");
        parent.$("#showBackupBtn").removeClass("disabled");

        parent.loadSnapshots();
    });
}
/*
 * Lock header (for subpages)
 * viewid		Parent page table id
 * scrollid		Parent page scrollbar container id
 * size			Keep the number of rows in the table when copying
 * divhead_id	Copy header id
 * tabid		Subpage table id
 */
function scroll(viewid, scrollid, size, divhead_id, tabid) {
    if (parent.$("#" + divhead_id).length > 0) {
        parent.$("#" + divhead_id).width($("#" + tabid).width());
        return;
    }

    var scroll = parent.document.getElementById(scrollid);

    var tb2 = parent.document.getElementById(viewid).cloneNode(true);

    var $table = $(parent.document.getElementById(viewid));
    if ($table.find("input[type='checkbox']").length > 0) {
        var id = $(tb2).find("input[type='checkbox']:first").attr("id");
        $table.find("input[type='checkbox']:first").removeAttr("id");
        $(tb2).find("input[type='checkbox']:first").attr("id", id);
    }

    for (var i = tb2.rows.length; i > size; i--) {
        tb2.deleteRow(size);
    }
    var top = parent.$("#" + viewid).offset().top;
    var left = parent.$("#" + viewid).offset().left;
    var bak = parent.document.createElement("div");

    scroll.appendChild(bak);
    bak.appendChild(tb2);
    bak.setAttribute("id", divhead_id);
    bak.style.position = "fixed";
    $(bak).css({
        "left": left,
        "top": top,
        width: $("#" + tabid).width(),
        backgroundColor: "#cfc",
        display: "block"
    });
    parent.$("#" + viewid).find("th").each(function () {
        this.innerHTML = "";
    });
}

function GetRequest() {
    var url = location.search;
    var theRequest = new Object();
    if (url.indexOf("?") != -1) {
        var str = url.substr(1);
        strs = str.split("&");
        for (var i = 0; i < strs.length; i++) {
            theRequest[strs[i].split("=")[0]] = unescape(strs[i].split("=")[1]);
        }
    }
    return theRequest;
}