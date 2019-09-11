
function TabModel() {
    this.thSel = "";
    this.tbSel = "";
    this.rowCount = 0;
    this.selCount = 0;
}
function eSelAll(tabModel) {
    initTabModel(tabModel);
    bindEvent(tabModel);
}
function initTabModel(tabModel) {
    tabModel.rowCount = $(tabModel.tbSel + " tr ").length;
    $.debug("initData rowCount: " + tabModel.rowCount);
    tabModel.selCount = $(tabModel.tbSel + " :checked ").length;
    $.debug("initData selCount: " + tabModel.selCount);
}
function bindEvent(tabModel) {
    $(tabModel.thSel).click(function () {
        $.debug("selAll click 前的selCount: " + tabModel.selCount);
        if ($(this).prop("checked")) {
            $(tbSel).prop("checked", true);
            tabModel.selCount = tabModel.rowCount;
        }
        else {
            $(tbSel).prop("checked", false);
            tabModel.selCount = 0;
        }
        $.debug("disAll click 后的selCount: " + tabModel.selCount);
    });
    $(tabModel.tbSel).click(function () {
        $.debug("click 前的selCount: " + tabModel.selCount);
        if ($(this).prop("checked")) {
            $.debug($(this).prop("id") + "选中....");
            tabModel.selCount++;
            if (tabModel.rowCount == tabModel.selCount)
            {
                $(thSel).prop("checked", true);
            }
        }
        else {
            $.debug($(this).prop("id") + "取消....");
            tabModel.selCount--;
            if (tabModel.rowCount > tabModel.selCount)
            {
                $(thSel).prop("checked", false);
            }
        }
        ;
        $.debug("click 后的selCount: " + tabModel.selCount);
    });
}

function a2t(tbsel, cloId, objs) {
    for (var i = 0; i < objs.length; i++) {
        var $lunRow = $(cloId).clone(true);
        $lunRow.attr("id", cloId + i);
        obj2rel($lunRow, objs[i]);
        $(tbsel).append($lunRow);
        $lunRow.show();
    }
}

function obj2rel($tr, obj) {
    $tr.children("td").each(function (num, td) {
        if (!isEmpObj($(td).attr("name"))) {
            // chg 20181101 : ie11 compatibility bug
            console.log($(td).attr("name"));
            if (isEmpObj(obj[$(td).attr("name")]) && obj[$(td).attr("name")] != false) {
                td.innerHTML = "";
                td.title = "";
            }
            else if ($(td).attr("name") == 'usedBy') {
                if (obj.usedType == "Datastore") {
                    url = "../../../assets/images/Datastore.png";
                    td.innerHTML = "<img src=" + url + ">" + "&nbsp;&nbsp;" + obj.usedBy;
                    td.title = obj.usedBy;
                } else if (obj.usedType == "VirtualMachine") {
                    url = "../../../assets/images/VirtualMachine.png";
                    td.innerHTML = "<img src=" + url + ">" + "&nbsp;&nbsp;" + obj.usedBy;
                    td.title = obj.usedBy;
                } else {
                    td.innerHTML = obj.usedBy;
                    td.title = obj.usedBy;
                }
            }
            else if ($(td).attr("name") == 'capacityUsage') {
                td.innerHTML = makeRateChart(obj.capacityUsage, $("#volCapUsage", parent.document));
                td.title = "Total Capacity:" + obj.capacity + " Free Capacity: " + obj.capacity;
            }
            else if ($(td).attr("name") == 'storagePoolUsage') {
                td.innerHTML = makeRateChart(obj.storagePoolUsage , $("#poolCapUsage", parent.document));
                td.title = "Total Capacity:" + obj.storagePoolCapactiy + " Free Capacity: " + obj.storagePoolFreeCap;
            }
            else if ($(td).attr("name") == 'status') {
                var status = obj.status;
                if (status.toLowerCase() == "normal") {
                    url = "../../../assets/images/lun_status_normal.png";
                    td.innerHTML = "<img src=" + url + ">"+ "&nbsp;&nbsp;" + status;
                    td.title = status;
                }
                else if (status.toLowerCase() == "faulty") {
                    url = "../../../assets/images/lun_status_unnormal.png";
                    td.innerHTML = "<img src=" + url + ">" + "&nbsp;&nbsp;" + status;
                    td.title = status;
                } else {

                }
            }
            else {
                td.innerHTML = obj[$(td).attr("name")];
                td.title = td.innerHTML;
            }
        }
    });
}

function t2a(tableId) {
    var objList = new Array();
    $(tableId).find("tbody tr").each(function (nun, tr) {
        var length = $(tr).find(":checked").length;
        if (length > 0) {
            var obj = r2o(tr);
            objList.push(obj);
        }
    });
    return objList;
}


function r2o(tr) {
    var obj = new Object();
    $(tr).children("td").each(
        function (num, td) {
            if (!isEmpObj($(td).attr("name"))) {
                obj[$(td).attr("name")] = td.innerHTML;
            }
        });
    return obj;
}