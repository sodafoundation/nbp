var ns = org_opensds_storage_devices;
var hostId = WEB_PLATFORM.getHostId();
var serverGuid = WEB_PLATFORM.getServerGuid();
var urlForLun = ns.webContextPath + "/rest/data/host/unmountableVolumeList/";
var cSnapshotUrl = ns.webContextPath + "/rest/data/host/snapshot";
var lunObj = new Object(), devObj = new Object();
var snapshot = new Object();
var snapshot_array = new Array();
var snapshots = new Array();
var toPage_lun = 1;
var pagesize_lun = 10;
var toPage_sanapshot = 1;
var pagesize_snapshot = 10;
var divhead_id_lun = "bak_lun";
var divhead_id_snapshot = "bak_snapshot";
var loadpage2_data_params = "";
var runNum = 0;

var filterType = "";
var filterValue = "";

//nfs
var lun_fs_flag = "lun";
var urlForFs = ns.webContextPath + "/rest/nfsdata/fs/";
var fsSnapshotUrl = ns.webContextPath + "/rest/nfsdata/fsSnapshot";
var dsNfsUrl = ns.webContextPath + "/rest/nfsdata/datastore";
var pager1Url = "";
var pager2Url = "";
var pager1ResultUrl = "";
var pager2ResultUrl = "";
var fsObj = new Object();

$(document).ready(function () {
    //makeHelp();
    loadLunsOrFs();
    bundleEvent();
});
/**
 * 初始化LUN或者file system数据
 */
function loadLunsOrFs() {
    $("#chk_all").prop("checked", false);
    $("#toggleLunFsBtn").prop("disabled", "disabled");
    lunObj = new Object(), devObj = new Object();
    fsObj = new Object();
    if ($("#" + divhead_id_lun).length > 0) {
        $("#" + divhead_id_lun).width($("#divMain").width() - 22);
    }
    $('#lunTabFrame').prop("src", "");
    $('#snapshotFrame').prop("src", "");
    $("#pager1").remove();
    $("#pager2").remove();
    $("#showBackupBtn").addClass("disabled");
    $("#showBackupBtn").prop("disabled", "disabled");
    $("#recoverBtn").addClass("disabled");
    $("#recoverBtn").prop("disabled", "disabled");
    $("#refreshSnapBtn").addClass("disabled");
    $("#refreshSnapBtn").prop("disabled", "disabled");
    $("#delSnapBtn").addClass("disabled");
    $("#delSnapBtn").prop("disabled", "disabled");
    $("#divLoadingSnapshot").hide();
    $("#divLoadingLun").show();

    if (lun_fs_flag == "lun") {
        pager1Url = urlForLun + "count/" + hostId + "?serverGuid=" + serverGuid;
        pager1ResultUrl = ns.webContextPath + "/resources/html/snapshot/lunTab.html";
    } else if (lun_fs_flag == "fs") {
        pager1Url = urlForFs + "count/" + hostId + "?serverGuid=" + serverGuid;
        pager1ResultUrl = ns.webContextPath + "/resources/html/snap shot/fsTab.html";
    }
    var url = pager1Url + "&filterType=" + filterType + "&filterValue=" + filterValue;
    $("#hostLunTab").bigPage({
        container: "pager1",
        ajaxData: {
            url: encodeURI(url),
            params: {
                loaddingId: "divLoadingLun",
                iframeId: "lunTabFrame",
                data_url: pager1ResultUrl,
                data_params: ''
            }
        },
        pageSize: pagesize_lun,
        toPage: toPage_lun,
        position: "down",
        callback: enableToggleBtn
    });
}

function enableToggleBtn() {
    $("#toggleLunFsBtn").prop("disabled", "");
}
/**
 * 加载快照数据
 */
function loadSnapshots() {

    if (lun_fs_flag == "lun") {
        loadpage2_data_params = "&lunId=" + lunObj.id + "&deviceId=" + devObj.id + "&t=" + new Date();
    } else if (lun_fs_flag == "fs") {
        loadpage2_data_params = "&fsId=" + fsObj.id + "&deviceId=" + devObj.id + "&t=" + new Date();
    }
    $("#chk_all").prop("checked", false);
    if (lunObj.id == "") {
        $("#divLoadingSnapshot").hide();
        return;
    }
    if ($("#" + divhead_id_snapshot).length > 0) {
        $("#" + divhead_id_snapshot).width($("#divMain").width() - 22);
    }
    snapshot = new Object();
    $("#recoverBtn").addClass("disabled");
    $("#delSnapBtn").addClass("disabled");
    $('#snapshotFrame').prop("src", "");
    $("#pager2").remove();
    $("#divLoadingSnapshot").show();

    if (lun_fs_flag == "lun") {
        pager2Url = ns.webContextPath + "/rest/data/host/snapshot/count?volumeId=" + lunObj.id + "&storageId=" + devObj.id + "&t=" + new Date();
        pager2ResultUrl = ns.webContextPath + "/resources/html/snapshot/lunSnapshotTab.html";
    } else if (lun_fs_flag == "fs") {
        pager2Url = ns.webContextPath + "/rest/nfsdata/fsSnapshot/count?fsId=" + fsObj.id + "&deviceId=" + devObj.id + "&t=" + new Date();
        pager2ResultUrl = ns.webContextPath + "/resources/html/snapshot/fsSnapshotTab.html";
    }
    $("#snapshotTab").bigPage({
        container: "pager2",
        ajaxData: {
            url: encodeURI(pager2Url),
            params: {
                loaddingId: "divLoadingSnapshot",
                iframeId: "snapshotFrame",
                data_url: pager2ResultUrl,
                data_params: loadpage2_data_params
            }
        },
        pageSize: pagesize_snapshot,
        toPage: toPage_sanapshot,
        position: "down",
        callback: function () {
            if ((lun_fs_flag == "lun" && isEmpObj(lunObj.id)) || (lun_fs_flag == "fs" && isEmpObj(fsObj.id))) {
                $("#snapshotFrame")[0].contentWindow.$("#snapshotTable").remove();
                $("#pager2").remove();
                return;
            }
        }
    });
}
/** 绑定事件处理 */
function bundleEvent() {
    //切换文件系统和块存储系统
    $("#toggleLunFsBtn").click(function () {

        if (lun_fs_flag == "lun") {
            lun_fs_flag = "fs";
            $("#toggleLunFsBtn").val("Lun");
            $("#lunTabBasicTitle").html("File System");
            $(".lunFsIdTh").attr("title", "File System Id");
            $(".lunFsIdDiv").html("FS Id");
            $(".lunTabWwnTh").hide();
            $(".lunTabWwnDiv").hide();
            $(".lunTabMappingTh").hide();
            $(".lunTabMappingDiv").hide();
            $(".spTabWwnTh").hide();
            $(".spTabWwnDiv").hide();
            $(".spRunningStatusTh").hide();
            $(".spRunningStatusDiv").hide();
            $(".capacityTh").hide();
            $(".capacityDiv").hide();
            $(".activedTh").attr("title", "Created");
            $(".activedDiv").html("Created");
            // $("#filterType option[value='ID']").remove();

            loadLunsOrFs();
        } else if (lun_fs_flag == "fs") {
            lun_fs_flag = "lun";
            $("#toggleLunFsBtn").val("File System");

            $("#lunTabBasicTitle").html("Lun");
            $(".lunFsIdTh").attr("title", "LUN ID");
            $(".lunFsIdDiv").html("LUN ID");
            $(".lunTabWwnTh").show();
            $(".lunTabWwnDiv").show();
            $(".lunTabMappingTh").show();
            $(".lunTabMappingDiv").show();
            $(".spTabWwnTh").show();
            $(".spTabWwnDiv").show();
            $(".spRunningStatusTh").show();
            $(".spRunningStatusDiv").show();
            $(".capacityTh").show();
            $(".capacityDiv").show();
            $(".activedTh").attr("title", "Activated");
            $(".activedDiv").html("Activated");
            // $("#filterType").append("<option value='ID'>ID</option>");

            loadLunsOrFs();
        }
    });

    //for search
    $("#btnSearch").click(function () {
        $("#btnSearch").prop("disabled", "disabled");
        $("#btnRefreshLUN").prop("disabled", "disabled");
        setTimeout(function () {
            $("#btnSearch").prop("disabled", "");
            $("#btnRefreshLUN").prop("disabled", "");
        }, 500);
        //点击过查询后,刷新操作才会过滤查询
        filterType = $("#filterType").val();
        if (filterType == "NAME" || filterType == "ID") {
            filterValue = trim($("#nameId_filterValue").val());
        }
        else if (filterType == "HEALTHSTATUS") {
            filterValue = $("#healthStatus_filterValue").val();
            if (filterValue == "NORMAL") {
            }
            else if (filterValue == "FAULT") {
            }
            else {
                filterValue = "";
            }
        }
        else if (filterType == "RUNNINGSTATUS") {
            filterValue = $("#runStatus_filterValue").val();
            if (filterValue == "ONLINE") {
            }
            else if (filterValue == "OFFLINE") {
            }
            else {
                filterValue = "";
            }
        }
        loadLunsOrFs();
    });

    $("#filterType").unbind();
    $("#filterType").bind("change", function () {
        $("#nameId_filterValue").val("");
        if ($('#filterType').val() == 'HEALTHSTATUS') {
            $("#nameId_filterValueDiv").hide();
            $("#runStatus_filterValue").hide();
            $("#healthStatus_filterValue").show();
            $("#healthStatus_filterValue").get(0).options[0].selected = true;
        }
        else if ($('#filterType').val() == 'RUNNINGSTATUS') {
            $("#nameId_filterValueDiv").hide();
            $("#healthStatus_filterValue").hide();
            $("#runStatus_filterValue").show();
            $("#runStatus_filterValue").get(0).options[0].selected = true;
        }
        else {
            $("#healthStatus_filterValue").hide();
            $("#runStatus_filterValue").hide();
            $("#nameId_filterValueDiv").show();
        }
    });

    // 刷新LUN列表
    $("#refreshLunBtn").click(function () {
        $("#hostLunTab tr:eq(0) th:eq(0)").width("5%");
        $("#hostLunTab tr:eq(0) th:eq(1)").width("10%");
        $("#hostLunTab tr:eq(0) th:eq(2)").width("10%");
        $("#hostLunTab tr:eq(0) th:eq(3)").width("10%");
        $("#hostLunTab tr:eq(0) th:eq(4)").width("10%");
        $("#hostLunTab tr:eq(0) th:eq(5)").width("5%");
        $("#hostLunTab tr:eq(0) th:eq(6)").width("10%");
        $("#hostLunTab tr:eq(0) th:eq(7)").width("20%");
        $("#hostLunTab tr:eq(0) th:eq(8)").width("10%");
        $("#hostLunTab tr:eq(0) th:eq(9)").width("10%");
        loadLunsOrFs();
    });
    // 刷新快照列表
    $("#refreshSnapBtn").click(function () {
        $("#snapshotTab tr:eq(0) th:eq(0)").width("4%");
        $("#snapshotTab tr:eq(0) th:eq(1)").width("22%");
        $("#snapshotTab tr:eq(0) th:eq(2)").width("10%");
        $("#snapshotTab tr:eq(0) th:eq(3)").width("10%");
        $("#snapshotTab tr:eq(0) th:eq(4)").width("10%");
        $("#snapshotTab tr:eq(0) th:eq(5)").width("24%");
        $("#snapshotTab tr:eq(0) th:eq(6)").width("20%");
        loadSnapshots();
    });
    // 取消对话框 取消删除 取消备份 取消还原
    $(".cancleBtn").click(function () {
        /*$("#overlay").css('visibility', 'hidden');*/
        $("#cSnapBox").css('visibility', 'hidden');
        $(".alertBox").hide();
        unlock();
    });
    // 删除对话框
    $("#delSnapBtn").click(function () {
        if (isEmpObj(devObj.id)) {
            $.debug("del snapshot check, data is null");
            return;
        }
        var id = "";
        snapshots = new Array();
        $snapshot_checked = $("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']:checked");
        $snapshot_checked.each(function (index, ele) {
            $tr = $(this).parent().parent();
            var snapshotObj = new Object();
            snapshotObj.id = trim($tr.find("[name='id']").text());
            snapshotObj.storageId = trim($tr.find("[name='storageId']").text());
            if (lun_fs_flag == "lun") {
                snapshotObj.parentId = trim($tr.find("[name='parentId']").text());
            } else if (lun_fs_flag == "fs") {
                snapshotObj.parentId = trim($tr.find("[name='fsId']").text());
            }

            if (isEmpObj(snapshotObj.id)) {
                $.debug("del snapshot check, data is null");
            }
            else {
                snapshots.push(snapshotObj);
            }
        });
        // TODO提示是否确认删除
        //$("#infWords").text("You are about to delete snapshot ("+snapshot.name+"). This operation cannot be undone. ");
        $("#infWords").text("You are about to delete snapshots. This operation cannot be undone. ");
        /*$("#overlay").css('visibility', 'visible');*/
        lock();
        $("#cSnapBox").css('visibility', 'visible');
        $(".alertBox").hide();
        $("#title").text("Delete Snapshot");
        $("#title").show();
        $("#delInfoDiv").show();
    });
    // 确认删除
    $("#confirBtn").click(function () {
        // 打开遮盖层
        /*$("#overlay").css('visibility', 'visible');*/
        $("#cSnapBox").css('visibility', 'hidden');
        $("#title").hide();
        $("#delInfoDiv").hide();
        /*$("#ingDiv").show();*/

        $("#sucType").val("refreshSnapBtn");
        $("#errType").val("refreshSnapBtn");
        // 发送请求信息
        var snapReq = "";
        if (lun_fs_flag == "lun") {
            snapReq = new req(cSnapshotUrl + "/" + devObj.id
                + "?hostId=" + hostId + "&serverGuid=" + serverGuid, JSON
                .stringify(snapshots));
        } else if (lun_fs_flag == "fs") {
            snapReq = new req(fsSnapshotUrl + "/" + devObj.id
                + "?hostId=" + hostId + "&serverGuid=" + serverGuid, JSON
                .stringify(snapshots));
        }
        snapReq.type = "DELETE";
        var snaphandler = new handler(function doSuccess(resp) {
            $(".alertBox").hide();
            $("#cSnapBox").css('visibility', 'visible');
            $("#title").text("Excute Result");
            $("#title").show();
            if (resp.data || resp.status == "ok") {
                $("#sucWords").text("Delete snapshot successfully.");
                $("#sucDiv").show();
            }
            else if (resp.msg || resp.status == "error") {
                $("#errWords").text(resp.msg);
                $("#errDiv").show();
            }
        }, function doFailed() {
            $(".alertBox").hide();
            $("#cSnapBox").css('visibility', 'visible');
            $("#title").text("Excute Result");
            $("#title").show();
            $("#errWords").text("Delete snapshot failed.");
            $("#errDiv").show();
        });
        sendMsg(snapReq, snaphandler);
    });
    // 备份窗口
    $("#showBackupBtn").click(function () {
        // 如果无选中
        if ((lun_fs_flag == "lun" && (isEmpObj(lunObj.id) || isEmpObj(lunObj.name))) ||
            lun_fs_flag == "fs" && (isEmpObj(fsObj.id) || isEmpObj(fsObj.name))) {
            $.debug("showBackupBtn check, data is null");
            return;
        }
        // 如果有选中
        var snapshotName = "";
        if (lun_fs_flag == "lun") {
            snapshotName = getSnapName(lunObj.name);// 将数据填充到对话框
        } else if (lun_fs_flag == "fs") {
            snapshotName = getSnapName(fsObj.name);// 将数据填充到对话框
        }
        $("#cSnapshotName").val(snapshotName);
        // 弹出对话框,将snapshot信息传递给对话框
        /*$("#overlay").css('visibility', 'visible');*/
        lock();
        $("#cSnapBox").css('visibility', 'visible');
        $(".alertBox").hide();
        $("#title").text("Create Snapshot");
        $("#title").show();
        $("#cSnapBoxContent").show();
    });
    // 确认备份
    $("#backupBtn").click(function () {
        $("#sucType").val("refreshSnapBtn");
        $("#errType").val("refreshSnapBtn");
        // 打开遮盖层
        /*$("#overlay").css('visibility', 'visible');*/
        $("#cSnapBox").css('visibility', 'hidden');
        $("#title").hide();
        $("#cSnapBoxContent").hide();
        /*$("#ingDiv").show();*/
        var snapshot = new Object();
        var lunReq = new Object();
        if (lun_fs_flag == "lun") {
            snapshot.parentId = lunObj.id;
            snapshot.name = $("#cSnapshotName").val();
            snapshot.storageId = devObj.id;
            lunReq = new req(cSnapshotUrl, JSON.stringify(snapshot));
            $.debug("name: " + snapshot.name + ".lunId: " + snapshot.lunId + ", deviceId: " + devObj.id);
        } else if (lun_fs_flag == "fs") {
            snapshot.parentId = fsObj.id;
            snapshot.name = $("#cSnapshotName").val();
            snapshot.storageId = devObj.id;
            lunReq = new req(fsSnapshotUrl, JSON.stringify(snapshot));
            $.debug("name: " + snapshot.name + ".fsId: " + snapshot.fsId + ", deviceId: " + devObj.id);
        }
        lunReq.type = "POST";
        var lunhandler = new handler(function doSuccess(resp) {
            if (resp.data || resp.status.toLowerCase() == "ok") {
                $(".alertBox").hide();
                $("#title").text("Excute Result");
                $("#title").show();
    			$("#sucWords").html(resp.msg);
                $("#sucWords").text("Create snapshot successful.");
                $("#sucDiv").show();
                $("#cSnapBox").css('visibility', 'visible');
            }
            else if (resp.msg || resp.status.toLowerCase() == "error") {
                $(".alertBox").hide();
                $("#title").text("Excute Result");
                $("#title").show();
                $("#errWords").html(resp.msg);
                $("#errDiv").show();
                $("#cSnapBox").css('visibility', 'visible');
            }
            else {
                $(".alertBox").hide();
                $("#title").text("Excute Result");
                $("#title").show();
                $("#errWords").html("Create snapshot failed.");
                $("#errDiv").show();
                $("#cSnapBox").css('visibility', 'visible');
            }
        }, function doFailed() {
            $(".alertBox").hide();
            $("#title").text("Excute Result");
            $("#title").show();
            $("#errWords").text("Create snapshot failed.");
            $("#errDiv").show();
            $("#cSnapBox").css('visibility', 'visible');
        });
        sendMsg(lunReq, lunhandler);
    });

    $("#sucOp").click(function () {
        $("#sucWords").text("");
        $("#title").text("");
        var sucType = $("#sucType").val();
        if (sucType == "refreshSnapBtn") {
            loadSnapshots();
        }
        else if (sucType == "refreshLunBtn") {
            loadLunsOrFs();
        }
        $(".alertBox").hide();
        $("#title").text("");
        /*$("#overlay").css('visibility', 'hidden');*/
        $("#cSnapBox").css('visibility', 'hidden');
        unlock();
    });
    $("#errOp").click(function () {
        $("#errWords").text("");
        $("#title").text("");
        var errType = $("#errType").val();
        if (errType == "refreshSnapBtn") {
            loadSnapshots();
        }
        else if (errType == "refreshLunBtn") {
            loadLunsOrFs();
        }
        $(".alertBox").hide();
        $("#title").text("");
        /*$("#overlay").css('visibility', 'hidden');*/
        $("#cSnapBox").css('visibility', 'hidden');
        unlock();
    });
    // 打开恢复对话框 - 恢复第一步
    $("#recoverBtn").click(function () {
        var used_tip = "";
        if (lun_fs_flag == "lun") {
            used_tip = "The source LUN is used by a datastore.Please unmount the datastore before rollbacking snapshot.";
        } else if (lun_fs_flag == "fs") {
            used_tip = "The source File System is used by a datastore. Click OK to unmount datastore for rollbacking snapshot.";
        }
        var scsiLunState_tip = "";
        scsiLunState_tip = "The source LUN is attached.please detach the source LUN before rollbacking snapshot.";
        var continue_tip = "";
        continue_tip = "The source LUN has been detached.You can rollback the snapshot.";
        if ((lun_fs_flag == "lun" ) || (lun_fs_flag == "fs")) {
            lock();
            $("#cSnapBox").css('visibility', 'visible');
            $(".alertBox").hide();
            $("#title").text("Before Rollbacking Snapshot");
            $("#title").show();

            $("#nextStep").addClass("disabled");
            $("#nextStep").prop("disabled", "disabled");
            $("#nextStep").css("background", "#57C7FF");
            if (lunObj.usedType == "Datastore" || fsObj.usedByStatus == "true") {
                $("#beforeRollback_tip").html(used_tip);
            }
            if (lun_fs_flag == "lun" && (lunObj.usedType != "Datastore" || isEmpObj(lunObj.usedType))) {
                $("#beforeRollback_tip").html(continue_tip);
                $("#nextStep").removeClass("disabled");
                $("#nextStep").prop("disabled", "");
                $("#nextStep").css("background", "#007cbb");
            } else if (lun_fs_flag == "fs") {
                $("#beforeRollback_tip").html(used_tip);
                $("#nextStep").removeClass("disabled");
                $("#nextStep").prop("disabled", "");
                $("#nextStep").css("background", "#007cbb");
            }
            $("#beforeRollback").show();
        }
        else {
            return;
        }
    });

    //点击ok按钮 - 恢复第二步
    $("#nextStep").click(function () {
        if (lun_fs_flag == "lun") {
            afterRemoveLunDs();
        } else if (lun_fs_flag == "fs") {
            $("#beforeRollback_tip").html("The datastore is removing, please wait for it done...");
            $("#nextStep").addClass("disabled");
            $("#nextStep").prop("disabled", "disabled");
            $("#nextStep").css("background", "#57C7FF");
            $("#cancelNext").addClass("disabled");
            $("#cancelNext").prop("disabled", "disabled");

            //为nfs增加卸载datastore的步骤,目的是为了保留datastore对应的文件系统
            var rmDsReq = new req(dsNfsUrl + "/" + fsObj.datastoreId + "?serverGuid=" + serverGuid + "&hostId=" + hostId, "");//是否需要转码

            rmDsReq.type = "DELETE";
            var rmDshandler = new handler(function doSuccess(resp) {
                if (resp.data) {
                    afterRemoveNfsDs();
                }
                else if (resp.msg || resp.status.toLowerCase() == "error") {
                    $(".alertBox").hide();
                    $("#title").text("Excute Result");
                    $("#title").show();
                    $("#errWords").html(resp.msg);
                    $("#errDiv").show();
                    $("#cSnapBox").css('visibility', 'visible');
                }
                else {
                    $(".alertBox").hide();
                    $("#title").text("Excute Result");
                    $("#title").show();
                    $("#errWords").html("Remove datastore failed.");
                    $("#errDiv").show();
                    $("#cSnapBox").css('visibility', 'visible');
                }
            }, function doFailed() {
                $(".alertBox").hide();
                $("#title").text("Excute Result");
                $("#title").show();
                $("#errWords").text("Remove datastore failed.");
                $("#errDiv").show();
                $("#cSnapBox").css('visibility', 'visible');
            });
            sendMsg(rmDsReq, rmDshandler);
        }


    });

    // 执行恢复 - 恢复第三步
    $("#rollbackBtn").click(function () {
        $("#sucType").val("refreshSnapBtn");
        $("#errType").val("refreshSnapBtn");
        if (lun_fs_flag == "lun") {
            // 打开遮盖层
            $("#cSnapBox").css('visibility', 'hidden');
            $("#title").hide();
            $("#rollbackContent").hide();
        } else if (lun_fs_flag == "fs") {

        }

        // 选择恢复速度
        var snapReq = new Object();
        if (lun_fs_flag == "lun") {
            snapReq = new req(cSnapshotUrl + "/" + devObj.id + "?rollbackSpeed=" + trim($("#rollbackSpeed").val()), JSON.stringify(snapshot));
        } else if (lun_fs_flag == "fs") {
            snapReq = new req(fsSnapshotUrl, JSON.stringify(snapshot));
        }
        snapReq.type = "PUT";
        var snaphandler = new handler(function doSuccess(resp) {
            if (lun_fs_flag == "lun") {
                $(".alertBox").hide();
                $("#cSnapBox").css('visibility', 'visible');
                $("#title").text("Excute Result");
                $("#title").show();
            }
            if (resp.data || resp.status.toLowerCase() == "ok") {
                if (lun_fs_flag == "lun") {
                    $("#sucWords").text("Recover snapshot successfully.");
                    $("#sucDiv").show();
                } else if (lun_fs_flag == "fs") {
                    //nfs还需把datastore挂载回来
                    $("#rollbackContentDiv").html("The datastore is mounting, please wait for it done...");
                    $("#rollbackBtn").addClass("disabled");
                    $("#rollbackBtn").prop("disabled", "disabled");
                    $("#cancelRB").addClass("disabled");
                    $("#cancelRB").prop("disabled", "disabled");
                    mountNfsDatastore();
                }
            }
            else if (resp.msg || resp.status.toLowerCase() == "error") {
                $("#errWords").text(resp.errorDesc);
                $("#errDiv").show();
            }
        }, function doFailed() {
            $(".alertBox").hide();
            $("#cSnapBox").css('visibility', 'visible');
            $("#title").text("Excute Result");
            $("#title").show();
            $("#errWords").text("Recovery snapshot failed.");
            $("#errDiv").show();
        });
        sendMsg(snapReq, snaphandler);
    });
    // 绑定快照名称输入框的校验
    $("#cSnapshotName").bind("input propertychange", function () {
        var value = $(this).val();//大于31位则不能输入,中文算3个字符
        var length = getLength(value);
        if (length > 31) {
            this.value = getByteVal(value, 31);
        }
        for (var index = 0; index < this.value.length; index++) {
            if (!(/^[a-zA-Z0-9-_.\u4e00-\u9fa5]$/.test(this.value.charAt(index)))) {
                this.value = this.value.substring(0, index);
            }
        }
        chkSnapName();
    }).bind("blur", function () {
        chkSnapName();
    });

    $("#nameId_filterValue").bind("input propertychange blur", function () {
        filterValue = $(this).val();
    });
}

//nfs恢复第二步执行成功的附加步骤
function afterRemoveLunDs() {
    $("#beforeRollback").hide();
    if (isEmpObj(devObj.id)) {
        return;
    }
    $snapshot_checked = $("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']:checked").first();
    $tr = $snapshot_checked.parent().parent();
    snapshot = new Object();
    snapshot.id = trim($tr.find("[name='id']").text());
    snapshot.storageId = devObj.id;
    snapshot.parentId = lunObj.id;
    snapshot.name = trim($tr.find("[name='name']").text());
    if (isEmpObj(snapshot.id)) {
        $.debug("del snapshot check, data is null");
        return;
    }
    $(".alertBox").hide();
    $("#rollbackSnap").text(snapshot.name);
    $("#title").text("Recover Snapshot");
    $("#title").show();
    $("#rollbackContent").show();
    if (lun_fs_flag == "lun") {
        $(".rollbackSpeedTr").show();
    } else if (lun_fs_flag == "fs") {
        $(".rollbackSpeedTr").hide();
        $("#rollbackContentDiv").html("You are about to restore the data on the source File system to the point in time when the\n\t\t\t\t\tsnapshot was created.");
    }
}

function afterRemoveNfsDs() {
    $("#beforeRollback").hide();
    if (isEmpObj(devObj.id)) {
        return;
    }
    $snapshot_checked = $("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']:checked").first();
    $tr = $snapshot_checked.parent().parent();
    snapshot = new Object();
    snapshot.id = trim($tr.find("[name='id']").text());
    snapshot.storageId = devObj.id;
    snapshot.fsId = fsObj.id;
    snapshot.name = trim($tr.find("[name='name']").text());
    if (isEmpObj(snapshot.id)) {
        $.debug("del snapshot check, data is null");
        return;
    }
    /*$("#overlay").css('visibility', 'visible');*/
    /*lock();
     $("#cSnapBox").css('visibility', 'visible');*/
    $(".alertBox").hide();
    $("#rollbackSnap").text(snapshot.name);
    $("#title").text("Recover Snapshot");
    $("#title").show();
    $("#rollbackContent").show();
    if (lun_fs_flag == "lun") {
        $(".rollbackSpeedTr").show();
    } else if (lun_fs_flag == "fs") {
        $(".rollbackSpeedTr").hide();
        $("#rollbackContentDiv").html("You are about to restore the data on the source File system to the point in time when the\n\t\t\t\t\tsnapshot was created.");
    }
}

//nfs恢复第三步执行成功后附加步骤
function mountNfsDatastore() {

    var nfsDatastoreMount = new Object();
    nfsDatastoreMount.localPath = fsObj.localPath;
    nfsDatastoreMount.remoteHost = fsObj.remoteHost;
    nfsDatastoreMount.remotePath = fsObj.remotePath;
    var nfsDsReq = new req(dsNfsUrl + "/" + fsObj.datastoreId + "?serverGuid=" + serverGuid + "&hostId=" + hostId, JSON.stringify(nfsDatastoreMount));

    nfsDsReq.type = "POST";
    var nfsDshandler = new handler(function doSuccess(resp) {
        $(".alertBox").hide();
        $("#cSnapBox").css('visibility', 'visible');
        $("#title").text("Excute Result");
        $("#title").show();
        if (resp.data) {
            $("#sucWords").text("Mount NFS datastore successfully.");
            $("#sucDiv").show();
        }
        else if (resp.errorCode) {
//          $("#errWords").text("Mount NFS datastore failed.");
            $("#errWords").text(resp.errorDesc);
            $("#errDiv").show();
        }
    }, function doFailed() {
        $(".alertBox").hide();
        $("#cSnapBox").css('visibility', 'visible');
        $("#title").text("Excute Result");
        $("#title").show();
        $("#errWords").text("Mount NFS datastore failed.");
        $("#errDiv").show();
    });
    sendMsg(nfsDsReq, nfsDshandler);
}

function init() {
    var table = $("#snapshotFrame")[0].contentWindow.$("#snapshotTable")[0];
    rowNum = table.rows.length;
    if (rowNum > 0) {
        setPageCheckBox(rowNum);
    }
}
/**
 * 设置页面多选框,rowNum为table行数
 * @param rowNum
 * @return
 */
function setPageCheckBox(rowNum) {
    var $singleChkbox = $("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']");
    $("#chk_all").unbind("click");
    $("#chk_all").click(function () {
        var chkAll = this.checked;
        $singleChkbox.each(function (i) {
            if (chkAll) {
                if (!this.checked) {
                    this.checked = true;
                }
            } else {
                if (this.checked) {
                    this.checked = false;
                }
            }
        });
        if (chkAll) {
            if (rowNum > 0) {
                $("#delSnapBtn").prop("disabled", "");
                $("#delSnapBtn").removeClass("disabled");
                $("#delSnapBtn .plugin_button_div").css("cursor", "pointer");
                if (rowNum == 1) {
                    if (lun_fs_flag == "lun") {
                        var lunRunningStatus = $("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']:checked").parent().parent().find("td[name='runningStatus']").text();
                        if (lunRunningStatus == "INACTIVATED") {
                            $("#recoverBtn").prop("disabled", "disabled");
                            $("#recoverBtn").addClass("disabled");
                            $("#recoverBtn .plugin_button_div").css("cursor", "default");
                        }
                        else {
                            $("#recoverBtn").prop("disabled", "");
                            $("#recoverBtn").removeClass("disabled");
                            $("#recoverBtn .plugin_button_div").css("cursor", "pointer");
                        }
                    } else if (lun_fs_flag == "fs") {
                        $("#recoverBtn").prop("disabled", "");
                        $("#recoverBtn").removeClass("disabled");
                        $("#recoverBtn .plugin_button_div").css("cursor", "pointer");
                    }

                }
                else {
                    $("#recoverBtn").prop("disabled", "disabled");
                    $("#recoverBtn").addClass("disabled");
                    $("#recoverBtn .plugin_button_div").css("cursor", "default");
                }
            }
        } else {
            $("#delSnapBtn").prop("disabled", "disabled");
            $("#recoverBtn").prop("disabled", "disabled");
            $("#delSnapBtn").addClass("disabled");
            $("#recoverBtn").addClass("disabled");
            $("#delSnapBtn .plugin_button_div").css("cursor", "default");
            $("#recoverBtn .plugin_button_div").css("cursor", "default");
        }

    });
    $singleChkbox.unbind("click");
    $singleChkbox.click(function () {
        var num = 0;

        $singleChkbox.each(function (i) {
            if (this.checked) {
                num++;
            }
        });
        if (num == rowNum) {
            if (!$("#chk_all")[0].checked) {
                $("#chk_all").prop("checked", "checked");
            }
        } else {
            $("#chk_all").prop("checked", "");
        }
        if (num > 0) {
            $("#delSnapBtn").prop("disabled", "");
            $("#delSnapBtn").removeClass("disabled");
            $("#delSnapBtn .plugin_button_div").css("cursor", "pointer");
            if (num == 1) {
                if (lun_fs_flag == "lun") {
                    var lunRunningStatus = $("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']:checked").parent().parent().find("td[name='runningStatus']").text();
                    if (lunRunningStatus == "INACTIVATED") {
                        $("#recoverBtn").prop("disabled", "disabled");
                        $("#recoverBtn").addClass("disabled");
                        $("#recoverBtn .plugin_button_div").css("cursor", "default");
                    }
                    else {
                        $("#recoverBtn").prop("disabled", "");
                        $("#recoverBtn").removeClass("disabled");
                        $("#recoverBtn .plugin_button_div").css("cursor", "pointer");
                    }
                } else if (lun_fs_flag == "fs") {
                    $("#recoverBtn").prop("disabled", "");
                    $("#recoverBtn").removeClass("disabled");
                    $("#recoverBtn .plugin_button_div").css("cursor", "pointer");
                }

            }
            else {
                $("#recoverBtn").prop("disabled", "disabled");
                $("#recoverBtn").addClass("disabled");
                $("#recoverBtn .plugin_button_div").css("cursor", "default");
            }
        } else {
            $("#delSnapBtn").prop("disabled", "disabled");
            $("#recoverBtn").prop("disabled", "disabled");
            $("#delSnapBtn").addClass("disabled");
            $("#recoverBtn").addClass("disabled");
            $("#delSnapBtn .plugin_button_div").css("cursor", "default");
            $("#recoverBtn .plugin_button_div").css("cursor", "default");
            $("#chk_all").prop("checked", "");
        }
    });
}

/**
 * 生成快照名称
 */
function getSnapName(lunName) {
    var cTime = new Date();
    var cMon = subTime("" + (cTime.getMonth() + 1));
    var cDate = subTime("" + cTime.getDate());
    var cHou = subTime("" + cTime.getHours());
    var cMin = subTime("" + cTime.getMinutes());
    var cSec = ("" + cTime.getSeconds());
    var suffix = ("" + cTime.getFullYear()).substring(2, 4) + cMon + cDate + cHou + cMin + cSec + cTime.getMilliseconds();// 名称后缀
    var snapshotName = lunName;
    if (snapshotName.length > 16) {
        snapshotName = snapshotName.substring(0, 16);
    }
    snapshotName = snapshotName + suffix;
    return snapshotName;
}
/* 补齐双位数 */
function subTime(tStr) {
    tStr = tStr.length == 1 ? "0" + tStr : tStr;
    return tStr;
}
/** 校验快照名称 */
function chkSnapName() {
    var sName = $("#cSnapshotName").val();
    if (isEmpObj(sName)) {
        $("#backupBtn").prop("disabled", true);
    }
    else {
        $("#backupBtn").prop("disabled", false);
    }
}
//锁屏
function lock() {
    $("#popupDiv").show();
    $("#ingDiv1").show();
}
//解屏 
function unlock() {
    $("#popupDiv").hide();
    $("#ingDiv1").hide();
}
function getLength(value) {
    var byteValLen = 0;
    var val = value.match(/./g);
    if (null == val) {
        return byteValLen;
    }
    for (var i = 0; i < val.length; i++) {
        if (val[i].match(/[^\x00-\xff]/ig) != null) {
            byteValLen += 3;
        } else {
            byteValLen += 1;
        }
    }
    return byteValLen;
}
function getByteVal(value, max) {
    var returnValue = "";
    var byteValLen = 0;
    var val = value.match(/./g);
    if (null == val) {
        return returnValue;
    }
    for (var i = 0; i < val.length; i++) {
        if (val[i].match(/[^\x00-\xff]/ig) != null) {
            byteValLen += 3;
        } else {
            byteValLen += 1;
        }
        if (byteValLen > max) {
            break;
        }
        returnValue += val[i];
    }
    return returnValue;
}
function trim(str) {
    return str.replace(/(^\s*)|(\s*$)/g, "");
}