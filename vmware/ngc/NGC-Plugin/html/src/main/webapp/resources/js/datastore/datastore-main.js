// COMMON
var KB_TO_BYTE = 1024;
var MB_TO_BYTE = 1024 * KB_TO_BYTE;
var GB_TO_BYTE = 1024 * MB_TO_BYTE;
var TB_TO_BYTE = 1024 * GB_TO_BYTE;
var TB_TO_GB = 1024;
var PB_TO_GB = 1024 * TB_TO_GB;
var dateTime = new Date();
var license = false;
var operationTitle = "<s:text name='storage.plugin.action.create.datastore.title'/>";
var currentPage;
var storType = "LUN";
var deviceIp = "";
var deviceId = "";
var deviceType = "";
var supportNFS = "";
var serverGuid = "";
var deviceName = "";
var storagePoolId = "";
var storagePoolName = "";
var vmfsVersion = "VMFS6";
// VMFS3 BLIOCK SIZE, DEFAULT 1
var blockSize = "1";
var lunName = "";
var produceDefaultName = true;

var lunDescription = "";
// TYPE :thin thick
var lunAllocType = "thick";
// capactity
var lunCapa = "";
var lunCapaUnit = "GB";
var useableCapa = "";
var initCapa = "";
var initCapaUnit = "GB";
var datastoreName = "";
var dsNameRepeatInfo = "";
var datastoreID = "";
var isExtentDataStor = false;
var storagePoolType = "";

// datastoreName is repeat?
var isDsNameRepeat = false;
var isLunInfoValide = false;
// window.external.SetTitle(operationTitle);

//for NFS -----------
var datastoreType = "";
var fileSystemName = "";
var fileSystemCapacity = "";
var fileSystemCapacityUnit = "GB";
var fileSystemAllocType = "thick";
var fileSystemDescription = "";
var sharePath = "";
var sharePathDescription = "";
var shareClientPermission = "readOnly";
var nasAddress = "";

var backgroundColor = "#efefef";
var focusColor = "#2b5480";
var type = "";
var isCreateDatastore = true;
var objectId = "";
var hostIdArray = new Array();
var hostServiceIPsToString = "";

$(document).ready(function () {

    hostId = WEB_PLATFORM.getHostId();
    serverGuid = WEB_PLATFORM.getServerGuid();
    type = GetQueryString("type");
    objectId = GetQueryString("objectId");
    init();

    if (type == "createDatastoreOrLun") {
        currentPage = 0;
        if (currentPage == 0) {
            $("#dg_top_0").css("display", "block");
            $("#preStep").prop("disabled", "disabled");
            $("#nextStep").prop("disabled", "disabled");
            $("#nextStep").css("background", "#57C7FF");
            $("#dg_main_left_0").css("background-color", focusColor);
            $("#dg_main_left_0").css("color", "white");
            $("#dg_main_left_0").css("border-radius", "5px");
            $("#mainFrame").attr("src", "datastoreAndLun_host_list.html?objectId=" + objectId + "&serverGuid=" + serverGuid);
        }
    } else {

        currentPage = 1;
        if (currentPage == 1) {
            $("#dg_top_1").css("display", "block");
            $("#preStep").prop("disabled", "disabled");
            $("#nextStep").prop("disabled", "disabled");
            $("#nextStep").css("background", "#57C7FF");
            $("#dg_main_left_1").css("background-color", focusColor);
            $("#dg_main_left_1").css("color", "white");
            $("#dg_main_left_1").css("border-radius", "5px");
            $("#mainFrame").attr("src", "datastore-p2-devicelist.html");
        }
    }
    // makeHelp();
    // shieldCombinationKey();
    $("#preStep").click(function () {
        // 如果按钮被灰化,直接返回
        if (this.disabled == true) {
            return;
        }

        var curPage = currentPage - 1;
        if (currentPage <= 3) {
            handlePreAction("#dg_top_" + currentPage, "#dg_top_" + curPage, "#dg_main_left_" + currentPage, "#dg_main_left_" + curPage, "#img_main_left_" + currentPage, "#img_main_left_" + curPage);
        } else {
            if (datastoreType == "vmfsDatastore") {
                //创建lun时的页面控制 data:2018-12-10 author:qwx615620
                if (isCreateDatastore == false && currentPage == 5) {
                    handlePreAction("#dg_top_" + 7, "#dg_top_" + 4, "#dg_main_left_" + 7, "#dg_main_left_" + 4, "#img_main_left_" + 7, "#img_main_left_" + 4);
                } else {
                    handlePreAction("#dg_top_" + currentPage, "#dg_top_" + curPage, "#dg_main_left_" + currentPage, "#dg_main_left_" + curPage, "#img_main_left_" + currentPage, "#img_main_left_" + curPage);
                }
            } else if (datastoreType == "nfsDatastore") {
                if (currentPage == 4) {
                    handlePreAction("#dg_top_nfs_" + currentPage, "#dg_top_" + curPage, "#dg_main_left_nfs_" + currentPage, "#dg_main_left_" + curPage, "#img_main_left_" + currentPage, "#img_main_left_" + curPage);
                } else {
                    handlePreAction("#dg_top_nfs_" + currentPage, "#dg_top_nfs_" + curPage, "#dg_main_left_nfs_" + currentPage, "#dg_main_left_nfs_" + curPage, "#img_main_left_nfs_" + currentPage, "#img_main_left_nfs_" + curPage);
                }
            }
        }

        currentPage--;
        if (currentPage == 0) {
            $("#mainFrame").attr("src", "datastoreAndLun_host_list.html?objectId=" + objectId + "&serverGuid=" + serverGuid);
            this.disabled = true;
        }
        if (currentPage == 1) {
            $("#mainFrame").attr("src", "datastore-p2-devicelist.html");
            if (type != "createDatastoreOrLun") {
                this.disabled = true;
            }
        }
        if (currentPage == 2) {
            $("#mainFrame").attr("src", "datastore-p3-chooseDatastoreType.html");
        }
        if (currentPage == 3) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p1-storagepool.html");
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p1-storagepool.html");
            }
        }
        if (currentPage == 4) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p3-lunInfo.html");
                $("#nextStep").val("NEXT");
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p2-fileSystemInfo.html");
            }
        }
        if (currentPage == 5) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p2-fileSystemVersion.html");
                $("#nextStep").val("NEXT");
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p3-nfsShareInfo.html");
                $("#nextStep").val("NEXT");
            }
        }
        if (currentPage == 6) {
            if (datastoreType == "vmfsDatastore" && (isCreateDatastore == true)) {
                $("#mainFrame").attr("src", "block/block-p4-datastoreName.html");
                $("#nextStep").val("CREATE");
            } else if (datastoreType == "vmfsDatastore" && (isCreateDatastore == false)) {
                $("#mainFrame").attr("src", "block/block-p3-lunInfo.html");
                $("#nextStep").val("CREATE");
            } else if (datastoreType == "nfsDatastore") {
                $("#mainFrame").attr("src", "nfs/nfs-p4-datastore.html?enterType=" + type);
                $("#nextStep").val("CREATE");
            }
        }
        if (currentPage == 7) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p5-summary.html?enterType=" + type);
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p5-summary.html?enterType=" + type);
            }
        }
    });
    $("#nextStep").click(function () {
        if (this.disabled == true) {
            return;
        }
        if (isCreateDatastore == false && ($("#nextStep").val() == "FINISH")) {
            $("#mainPage").css("display", "none");
            createDatastore();
            WEB_PLATFORM.closeDialog();
            releaseMemory();
            return;
        }
        this.disabled = true;
        var curPage = currentPage + 1;
        if (currentPage < 3) {
            handleNextAction("#dg_top_" + currentPage, "#dg_top_" + curPage, "#dg_main_left_" + currentPage, "#dg_main_left_" + curPage, "#img_main_left_" + currentPage, "#img_main_left_" + curPage);
        } else {
            if (datastoreType == "vmfsDatastore") {
                if (isCreateDatastore == false && currentPage == 4) {
                    handleNextAction("#dg_top_" + currentPage, "#dg_top_" + 7, "#dg_main_left_" + currentPage, "#dg_main_left_" + 7, "#img_main_left_" + currentPage, "#img_main_left_" + 7);
                    $("#mainFrame").attr("src", "block/block-p5-summary.html?enterType=" + type);
                    $("#nextStep").val("FINISH");
                    this.disabled = false;
                    currentPage++;
                    return;
                } else {
                    handleNextAction("#dg_top_" + currentPage, "#dg_top_" + curPage, "#dg_main_left_" + currentPage, "#dg_main_left_" + curPage, "#img_main_left_" + currentPage, "#img_main_left_" + curPage);
                }
            } else if (datastoreType == "nfsDatastore") {
                if (currentPage == 3) {
                    handleNextAction("#dg_top_" + currentPage, "#dg_top_nfs_" + curPage, "#dg_main_left_" + currentPage, "#dg_main_left_nfs_" + curPage, "#img_main_left_" + currentPage, "#img_main_left_" + curPage);
                } else {
                    handleNextAction("#dg_top_nfs_" + currentPage, "#dg_top_nfs_" + curPage, "#dg_main_left_nfs_" + currentPage, "#dg_main_left_nfs_" + curPage, "#img_main_left_nfs_" + currentPage, "#img_main_left_nfs_" + curPage);
                }
            }
        }
        currentPage++;
        $("#preStep").prop("disabled", "");
        if (currentPage == 0) {
            $("#mainFrame").attr("src", "datastoreAndLun_host_list.html?objectId=" + objectId + "&serverGuid=" + serverGuid);
        }
        if (currentPage == 1) {
            $("#mainFrame").attr("src", "datastore-p2-devicelist.html");
            if (type == "createDatastoreOrLun") {
                this.disabled = true;
            }
        }
        if (currentPage == 2) {
            $("#mainFrame").attr("src", "datastore-p3-chooseDatastoreType.html?enterType=" + type);
            this.disabled = true;
        }
        if (currentPage == 3) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p1-storagepool.html");
                this.disabled = true;
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p1-storagepool.html");
                this.disabled = true;
            }
        }
        if (currentPage == 4) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p3-lunInfo.html");
                this.disabled = true;
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p2-fileSystemInfo.html");
                this.disabled = true;
            }
        }
        if (currentPage == 5) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p2-fileSystemVersion.html");
                $("#nextStep").val("NEXT");
                this.disabled = true;
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p3-nfsShareInfo.html");
                $("#nextStep").val("NEXT");
                this.disabled = true;
            }
        }
        if (currentPage == 6) {
            if (datastoreType == "vmfsDatastore" && (isCreateDatastore == true)) {
                $("#mainFrame").attr("src", "block/block-p4-datastoreName.html");
                $("#nextStep").val("CREATE");
                this.disabled = true;
            } else if (datastoreType == "vmfsDatastore" && (isCreateDatastore == false)) {
                $("#mainFrame").attr("src", "block/block-p5-summary.html?enterType=" + type);
                $("#nextStep").val("FINISH");
                this.disabled = false;
            } else if (datastoreType == "nfsDatastore") {
                $("#mainFrame").attr("src", "nfs/nfs-p4-datastore.html");
                $("#nextStep").val("CREATE");
                this.disabled = true;
            }
        }
        if (currentPage == 7) {
            if (datastoreType == "vmfsDatastore") {
                $("#mainFrame").attr("src", "block/block-p5-summary.html?enterType=" + type);
                $("#nextStep").val("FINISH");
                this.disabled = false;
            } else {
                $("#mainFrame").attr("src", "nfs/nfs-p5-summary.html?enterType=" + type);
                $("#nextStep").val("FINISH");
                this.disabled = false;
            }
        }
        if (currentPage == 8) {
            $("#mainPage").css("display", "none");
            createDatastore();
            WEB_PLATFORM.closeDialog();
            releaseMemory();
            return;
        }

    });
    $("#cancel").click(function () {
        WEB_PLATFORM.closeDialog();
        releaseMemory();
    });
    $(window).resize(function () {
        init();
    });
});

function GetQueryString(name) {
    var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)");
    var r = window.location.search.substr(1).match(reg);
    if (r != null)return unescape(r[2]);
    return null;
}

function handlePreAction(lastTopDivId, curTopDivId, lastLeftDivId, curLeftDivId, lastLeftImgId, curLeftImgId) {
    $(lastTopDivId).css("display", "none");
    $(lastLeftDivId).css("font-weight", "normal");
    $(lastLeftDivId).css("color", "#7f7f7f");
    $(lastLeftDivId).css("background-color", backgroundColor);
    $(lastLeftDivId).css("text-decoration", "");
    $(lastLeftImgId).css("display", "none");

    $(curTopDivId).css("display", "block");

    $(curLeftDivId).css("background-color", focusColor);
    $(curLeftDivId).css("color", "white");
    $(curLeftDivId).css("border-radius", "5px");
    $(curLeftImgId).css("display", "none");
}

function handleNextAction(lastTopDivId, curTopDivId, lastLeftDivId, curLeftDivId, lastLeftImgId, curLeftImgId) {
    $(lastTopDivId).css("display", "none");
    $(lastLeftDivId).css("font-weight", "normal");
    $(lastLeftDivId).css("background-color", backgroundColor);
    $(lastLeftDivId).css("color", "#000");
    $(lastLeftImgId).css("display", "block");
    //$(lastLeftDivId).css("text-decoration", "underline");

    $(curTopDivId).css("display", "block");
    $(curLeftDivId).css("background-color", "#2b5480");
    $(curLeftDivId).css("color", "white");
    $(curLeftDivId).css("border-radius", "5px");

}

function releaseMemory() {
    KB_TO_BYTE = null;
    MB_TO_BYTE = null;
    GB_TO_BYTE = null;
    TB_TO_BYTE = null;
    TB_TO_GB = null;
    PB_TO_GB = null;
    dateTime = null;
    license = null;
    operationTitle = null;
    currentPage = null;
    storType = null;
    deviceIp = null;
    deviceId = null;
    deviceType = null;
    deviceName = null;
    storagePoolId = null;
    storagePoolName = null;
    vmfsVersion = null;
    blockSize = null;
    lunName = null;
    produceDefaultName = null;
    lunDescription = null;
    lunAllocType = null;
    lunCapa = null;
    lunCapaUnit = null;
    useableCapa = null;
    initCapa = null;
    initCapaUnit = null;
    datastoreName = null;
    dsNameRepeatInfo = null;
    isDsNameRepeat = null;
    isLunInfoValide = null;
    profileID = null;
    datastoreID = null;
    hostIdArray = null;
    var ifr = document.getElementById('mainFrame');
    ifr.parentNode.removeChild(ifr);
    ifr = null;
    setTimeout("CollectGarbage();", 1);
}

function createDatastore() {
    lunCapa = lunCapa + lunCapaUnit;
    var createDatastoreInfo = new Object();
    createDatastoreInfo.name = datastoreName;
    createDatastoreInfo.type = datastoreType;
    createDatastoreInfo.isCreateDatastore = isCreateDatastore;
    if (datastoreType == "vmfsDatastore") {
        createDatastoreInfo.vmfsVersion = vmfsVersion;
        var volumeList = new Array();
        var volumeInfo = new Object();
        volumeInfo.name = lunName;
        volumeInfo.description = lunDescription;
        volumeInfo.capacity = lunCapa;
        volumeInfo.type = lunAllocType;
        volumeInfo.storageType = deviceType;
        volumeInfo.storageId = deviceId;
        volumeInfo.storagePoolId = storagePoolId;
        volumeList[0] = volumeInfo;
        createDatastoreInfo.volumeInfos = volumeList;
    }

    /*createDatastoreInfo.deviceType = deviceType;
     createDatastoreInfo.deviceId = deviceId;
     createDatastoreInfo.datastoreName = datastoreName;
     createDatastoreInfo.storagePoolId = storagePoolId;
     createDatastoreInfo.datastoreType = datastoreType;
     createDatastoreInfo.isCreateDatastore = isCreateDatastore;
     if(datastoreType == "vmfsDatastore") {
     createDatastoreInfo.vmfsVersion = vmfsVersion;
     createDatastoreInfo.lunName = lunName;
     createDatastoreInfo.lunDescription = lunDescription;
     createDatastoreInfo.lunCapacity = lunCapa;
     createDatastoreInfo.allocType = lunAllocType;
     }*/ else {
        //for nfs
        createDatastoreInfo.fileSystemName = fileSystemName;
        createDatastoreInfo.allocType = fileSystemAllocType;
        createDatastoreInfo.fileSystemCapacity = getCapaValue(fileSystemCapacity + fileSystemCapacityUnit);
        createDatastoreInfo.fileSystemDescription = fileSystemDescription;
        createDatastoreInfo.sharePath = sharePath;
        createDatastoreInfo.sharePathDescription = sharePathDescription;
        createDatastoreInfo.shareClientPermission = shareClientPermission;
        createDatastoreInfo.nasAddress = nasAddress;
        createDatastoreInfo.serviceIps = hostServiceIPsToString;
    }
    var actionUid = "org.opensds.ngc.action.createDatastore";
    var postJson = JSON.stringify(createDatastoreInfo);
    if (type == "createDatastoreOrLun") {
        url =  "/rest/datastore/create?actionUid=" + actionUid + "&objectId=" + hostIdArray + "&serverGuid=" + serverGuid;
    } else {
        url =   "/rest/datastore/create?actionUid=" + actionUid + "&objectId=" + hostId + "&serverGuid=" + serverGuid;
    }
    var ns = org_opensds_storage_devices;
    var requsetURL = ns.webContextPath + url;
    WEB_PLATFORM.callActionsController(requsetURL, postJson);
   /* $.ajax({
        async:true,
        contentType: "application/x-www-form-urlencoded;charset=UTF-8",
        type: "POST",
        url: encodeURI(requsetURL),
        data: postJson,
        dataType: 'json',
        success: function (data) {
            console.log(data);
        }, error: function (XMLHttpRequest, textStatus, errorThrown) {
            alert(XMLHttpRequest.status);
            alert(XMLHttpRequest.readyState);
            alert(textStatus);
        }
    })*/
}

function init() {
    $("#top").height(60);
    var height = $("body").height();
    var topHeight = $("#top").height();
    var filterHeight = $(".filterDiv").height() * 3;
    var bottomHeight = $("#dg_bottom").height();
    $("#main").height(365);
}

function nameRepeat() {
    $("#nextStep").prop("disabled", "disabled");
    $("#nextStep").css("background", "#57C7FF")
    $("#error", document.frames('mainIFrame').document).css("display", "none");
    var url = contextPath + "/action/datastore/judgeNameRepeat.action?time=" + dateTime.getTime();
    $
        .ajax({
            contentType: "application/x-www-form-urlencoded;charset=UTF-8",
            url: encodeURI(url),
            async: false,
            data: {
                datastoreName: datastoreName
            },
            success: function (resp) {
                var data = resp.data;
                if (data == "2") {
                    isDsNameRepeat = false;
                } else if (data == "1") {
                    $("#pointOutInfo",
                        document.frames('mainIFrame').document)
                        .text(
                            "<s:text name='storage.plugin.action.create.datastore.name.repeat'/>");
                    dsNameRepeatInfo = "<s:text name='storage.plugin.action.create.datastore.name.repeat'/>";
                    isDsNameRepeat = true;
                } else if (data == "0") {
                    $("#pointOutInfo",
                        document.frames('mainIFrame').document)
                        .text(
                            "<s:text name='storage.plugin.action.create.datastore.vcenter.break'/>");
                    dsNameRepeatInfo = "<s:text name='storage.plugin.action.create.datastore.vcenter.break'/>";
                    isDsNameRepeat = true;
                } else {
                    $("#pointOutInfo",
                        document.frames('mainIFrame').document)
                        .text(
                            "<s:text name='storage.plugin.action.create.datastore.name.too.long'/>");
                    dsNameRepeatInfo = "<s:text name='storage.plugin.action.create.datastore.name.too.long'/>";
                    isDsNameRepeat = true;
                }
            },
            complete: function (XHR, TS) {
                XHR = null;
            }
        });
}

function getCapaValue(capa) {
    var capaValue = parseFloat(capa);
    if (capa == "infinite") {
        return 64 * 1024;
    } else if (capa.indexOf("PB") > 0) {
        capaValue *= PB_TO_GB;
    } else if (capa.indexOf("TB") > 0) {
        capaValue *= TB_TO_GB;
    } else if (capa.indexOf("GB") <= 0) {
        capaValue = 0;
    }
    capaValue = parseInt(capaValue);
    return capaValue;
}

function getMaxCapaShow(capa) {
    if (capa.indexOf("GB") > 0) {
        return parseInt(capa) + " GB";
    }
    return capa;
}

function judgeLunInfo() {
    var capa_range = "<s:text name='storage.plugin.action.create.datastore.lun.capa.range'/>";
    var init_capa_range = "<s:text name='storage.plugin.action.create.datastore.init.capa.range'/>";

    var MAX_LUN_CAPA = 64 * 1024;

    var max_capa_value = "";
    var max_capa_show = "";

    var max_init_capa_value = "";
    var max_init_capa_show = "";

    var lunCapaValue = getCapaValue((lunCapa + lunCapaUnit));

    var initCapaValue = getCapaValue((initCapa + initCapaUnit));

    var usedCapaValue = getCapaValue(useableCapa);


    if (usedCapaValue >= MAX_LUN_CAPA) {
        max_capa_value = MAX_LUN_CAPA;
        max_capa_show = "64 TB";
    } else {
        max_capa_value = usedCapaValue;
        max_capa_show = getMaxCapaShow(useableCapa);
    }

    if (usedCapaValue >= lunCapaValue) {
        max_init_capa_value = lunCapaValue;
        max_init_capa_show = lunCapa + " " + lunCapaUnit;
    } else {
        max_init_capa_value = usedCapaValue;
        max_init_capa_show = getMaxCapaShow(useableCapa);
    }

    if (lunAllocType == "thick") {

        if (2 > lunCapaValue || lunCapaValue > max_capa_value) {
            $("#lunCapaRange", document.frames('mainIFrame').document).text(
                capa_range.replace("[xxx]", max_capa_show));
            $("#lunCapaRange", document.frames('mainIFrame').document).css(
                "display", "block");
            $("#lunCapaRangeImage", document.frames('mainIFrame').document)
                .css("display", "block");
            $("#txtCapa", document.frames('mainIFrame').document).addClass(
                "focus");
            $("#txtInitCapa", document.frames('mainIFrame').document)
                .removeClass("focus");
            return false;
        } else {
            $("#lunCapaRange", document.frames('mainIFrame').document).css(
                "display", "none");
            $("#lunCapaRangeImage", document.frames('mainIFrame').document)
                .css("display", "none");
            $("#txtCapa", document.frames('mainIFrame').document).removeClass(
                "focus");
            $("#txtInitCapa", document.frames('mainIFrame').document)
                .removeClass("focus");
            return true;
        }
    } else {

        if (2 > lunCapaValue || lunCapaValue > MAX_LUN_CAPA) {
            $("#lunCapaRange", document.frames('mainIFrame').document).text(
                capa_range.replace("[xxx]", "64 TB"));
            $("#lunCapaRange", document.frames('mainIFrame').document).css(
                "display", "block");
            $("#lunCapaRangeImage", document.frames('mainIFrame').document)
                .css("display", "block");
            $("#txtCapa", document.frames('mainIFrame').document).addClass(
                "focus");
            $("#txtInitCapa", document.frames('mainIFrame').document)
                .removeClass("focus");
            $("#initCapaRange", document.frames('mainIFrame').document).css(
                "display", "none");
            $("#initCapaRangeImage", document.frames('mainIFrame').document)
                .css("display", "none");
            return false;
        } else {
            $("#lunCapaRange", document.frames('mainIFrame').document).css(
                "display", "none");
            $("#lunCapaRangeImage", document.frames('mainIFrame').document)
                .css("display", "none");
            $("#txtCapa", document.frames('mainIFrame').document).removeClass(
                "focus");
        }

        if (1 > initCapaValue || initCapaValue > max_init_capa_value) {
            $("#initCapaRange", document.frames('mainIFrame').document).text(
                init_capa_range.replace("[xxx]", max_init_capa_show));
            $("#initCapaRange", document.frames('mainIFrame').document).css(
                "display", "block");
            $("#initCapaRangeImage", document.frames('mainIFrame').document)
                .css("display", "block");
            $("#txtInitCapa", document.frames('mainIFrame').document).addClass(
                "focus");
            return false;
        } else {
            $("#initCapaRange", document.frames('mainIFrame').document).css(
                "display", "none");
            $("#initCapaRangeImage", document.frames('mainIFrame').document)
                .css("display", "none");
            $("#txtInitCapa", document.frames('mainIFrame').document)
                .removeClass("focus");
            return true;
        }
    }
}