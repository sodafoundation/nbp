var hostId = parent.hostId;
var deviceId = parent.deviceId;
var serverGuid = parent.serverGuid;
var filterType = parent.filterType;
var filterValue = parent.filterValue;
var maxNum = 50;

$(document).ready(function() {
	initData();

});

function initData() {
	var request = new Object();
	request = GetRequest();
	var ns = org_opensds_storage_devices;
	parent.$("#divLoadingLun").show();
	parent.$("#chk_all").prop("checked", false);
	parent.$("#chk_all").attr("disabled", "disabled");
	var url = ns.webContextPath + "/rest/data/host/mountableVolumeList/" + hostId + "?deviceId=" + deviceId + "&serverGuid=" + serverGuid +
		"&filterType=" + filterType + "&filterValue=" + filterValue + "&start=" + request["start"] + "&count=" + request["pagesize"] + "&t=" + new Date();
	var lunReq = new req(url, "");
	var lunhandler = new handler(function doSuccess(resp) {
		if(resp.msg) {
			parent.$("#diverrorLUN").text(resp.msg).show();
			parent.$("#divLoadingMappedLUN").hide();
			return;
		}
		var arr = resp.data;
		for(var i = 0; i < arr.length; i++) {
			var $lunRow = $("#cloneLun").clone(true);
			$lunRow.attr("id", "#cloneLun" + i);
			$lunRow.children("td").each(function(num, td) {
				if(isEmpObj($(td).attr("name"))) {

				}
				else if($(td).attr("name") == 'check') {
					td.innerHTML = "<input type='checkbox' id=\'diveChbox_" + i + "\'style='vertical-align:middle;'/>";
				}
				else if($(td).attr("name") == 'hideDeviceId_') {
					td.id = "hideDeviceId_" + i;
					td.value = arr[i]["storageId"];
				}
				else if($(td).attr("name") == 'hideLunId_') {
					td.id = "hideLunId_" + i;
					td.value = arr[i]["id"];
				}
                else if ($(td).attr("name") == 'capacityUsage') {
                    td.innerHTML = makeRateChart(arr[i]["capacityUsage"], $("#volCapUsage", parent.document));
                    td.title = "Total Capacity:" + arr[i]["capacity"] + " Free Capacity: " + arr[i]["freeCapacity"];
                }
                else if ($(td).attr("name") == 'poolCapUsage') {
                    td.innerHTML = makeRateChart(arr[i]["storagePoolUsage"], $("#poolCapUsage", parent.document));
                    td.title = "Total Capacity:" + arr[i]["storagePoolCapactiy"] + " Free Capacity: " + arr[i]["storagePoolFreeCap"];
                }
                else if ($(td).attr("name") == 'status') {
                    var status = arr[i]["status"];
                    if (status == "Normal") {
                        url = "../../../assets/images/lun_status_normal.png";
                        td.innerHTML = "<img src=" + url + ">"
                        td.title = status;
                    }
                    else if (status == "Faulty") {
                        url = "../../../assets/images/lun_status_unnormal.png";
                        td.innerHTML = "<img src=" + url + ">"
                        td.title = status;
                    }
                }
				else {
					td.innerHTML = arr[i][$(td).attr("name")];
					td.title = td.innerHTML;
				}
			});
			$("#hostLunTbody").append($lunRow);
			$lunRow.show();
		}
		scroll("mappedLunList", "mappedLunListDiv", 1, parent.divhead_id_lun, "hostLunTable");
		loaclInit();
		parent.$("#divLoadingMappedLUN").hide();
		parent.$("#chk_all").attr("disabled", false);

	}, function doFailed() {
		parent.$("#divLoadingMappedLUN").hide();
	});
	sendMsg(lunReq, lunhandler);
}
/**
 * 仅仅提供单选按钮,并不提供批量快照/备份功能, 如果提供也不再通过checkbox的方式
 */
function bindEvent() {
	$("#hostLunTable tbody tr").bind("click", function(event) { // 行的点击事件

		$("#hostLunTable tbody tr td").css("background-color", "#FFFFFF"); // 删除其他选中行的背景样式
		$(this).find('td').each(function(i) {
			$(this).css("background-color", "#abcefc");
		});

		parent.lunObj.id = $(this).find("[name='id']").text();
		parent.lunObj.name = $(this).find("[name='name']").text();
		parent.lunObj.status = $(this).find("[name='status']").text();
		parent.lunObj.usedByStatus = $(this).find("[name='usedByStatus']").text();
		parent.devObj.id = $(this).find("[name='serialNumber']").text();

		parent.$("#refreshSnapBtn").prop("disabled", "");
		parent.$("#refreshSnapBtn").removeClass("disabled");
		parent.$("#refreshLunBtn").prop("disabled", "");

		parent.$("#showBackupBtn").prop("disabled", "");
		parent.$("#showBackupBtn").removeClass("disabled");

		parent.loadSnapshots();
	});
}
/*
 * 锁定表头（用于子页面）
 * viewid		父页面table id
 * scrollid		父页面滚动条容器id
 * size			copy时保留表格的行数
 * divhead_id	copy的表头id
 * tabid		子页面表格id
 */
function scroll(viewid, scrollid, size, divhead_id, tabid) {
	if(parent.$("#" + divhead_id).length > 0) {
		parent.$("#" + divhead_id).width($("#" + tabid).width());
		return;
	}

	var scroll = parent.document.getElementById(scrollid);
	var tb2 = parent.document.getElementById(viewid).cloneNode(true);

	var $table = $(parent.document.getElementById(viewid));
	if($table.find("input[type='checkbox']").length > 0) {
		var id = $(tb2).find("input[type='checkbox']:first").attr("id");
		$table.find("input[type='checkbox']:first").removeAttr("id");
		$(tb2).find("input[type='checkbox']:first").attr("id", id);
	}

	for(var i = tb2.rows.length; i > size; i--) {

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
	parent.$("#" + viewid).find("th").each(function() {
		this.innerHTML = "";
	});
}

function GetRequest() {
	var url = location.search; //获取url中"?"符后的字串
	var theRequest = new Object();
	if(url.indexOf("?") != -1) {
		var str = url.substr(1);
		strs = str.split("&");
		for(var i = 0; i < strs.length; i++) {
			theRequest[strs[i].split("=")[0]] = unescape(strs[i].split("=")[1]);
		}
	}
	return theRequest;
}
/**
 * 获取LUN的ID列表
 */
function getLunIds() {
	var ids = [];
	$.each($("input[id^='diveChbox_']"), function(i) {
		if(this.checked) {
			var id = this.id.split("_")[1];
			ids.push($("#hideLunId_" + id).val());
		}
	});
	return ids;
}

/**
 * 获取与LUN对应的阵列的ID列表
 */
function getDeviceIdsInMount() {
	var ids = [];
	$("input[id^='diveChbox_']").each(function(i) {
		//$.each($("input[id^='diveChbox_']"), function(i) {
		if(this.checked) {
			var id = this.id.split("_")[1];
			ids.push($("#hideDeviceId_" + id).val());
		}
	});
	return ids;
}

/**
 * 获取列表行数
 */
function loaclInit() {
	var table = document.getElementById("hostLunTable");
	var rowNum = table.rows.length - 1;
	if(rowNum > 0) {
        localSetPageCheckBox(rowNum);
	}
}

/**
 * 设置页面多选框,rowNum为table行数
 * @param rowNum
 * @return
 */
function localSetPageCheckBox(rowNum) {
	parent.$("#chk_all").unbind("click");
	parent.$("#chk_all").click(function() {
		var message = "The number of LUNs you choose can not be more than " + maxNum + "!(The default selection of the first " + maxNum + " data)";
		var flag = moreThanMaxNum(rowNum, 0, message);
		if(flag) {
			this.checked = false;
			//默认选中前面maxNum条数据
			$("input[id^='diveChbox_']").each(function(i, n) {
				if(i < maxNum) {
					this.checked = true;
				} else {
					this.checked = false;
				}
			});
			return;
		} else {
			if(rowNum > maxNum) {
				return;
			}
		}
		var message = "The number of LUNs you choose can not <br />be more than " + maxNum + "!";
		var flag = moreThanMaxNum(rowNum, 0, message);
		if(flag) {
			this.checked = false;
			return;
		} else {
			if(rowNum > maxNum) {
				return;
			}
		}

		var chkAll = this.checked;
		if(chkAll) {
			if(rowNum > 0) {
				parent.$("#btnMount").prop("disabled", "");
				parent.$("#btnMount .plugin_button_div").css("cursor", "pointer");
			}
		} else {
			parent.$("#btnMount").prop("disabled", "disabled");
			parent.$("#btnMount .plugin_button_div").css("cursor", "default");
		}
		$("input[id^='diveChbox_']").each(function(i) {
			if(chkAll) {
				if(!this.checked) {
					this.checked = true;
				}
			} else {
				if(this.checked) {
					this.checked = false;
				}
			}
		});
	});

	$("input[id^='diveChbox_']").unbind("click");
	$("input[id^='diveChbox_']").click(function() {
		var message = "The number of LUNs you choose can not <br />be more than " + maxNum + "!";
		var num = 0;

		$("input[id^='diveChbox_']").each(function(i) {
			if(this.checked) {
				num++;
			}
		});
		var flag = moreThanMaxNum(rowNum, num, message);
		if(flag) {
			this.checked = false;
			return;
		}
		if(num == rowNum) {
			if(!parent.$("#chk_all")[0].checked) {
				parent.$("#chk_all").prop("checked", "checked");
			}
		} else {
			parent.$("#chk_all").prop("checked", "");
		}
		if(num > 0) {
			parent.$("#btnMount").prop("disabled", "");
			parent.$("#btnMount .plugin_button_div").css("cursor", "pointer");
		} else {
			parent.$("#btnMount").prop("disabled", "disabled");
			parent.$("#btnMount .plugin_button_div").css("cursor", "default");
			parent.$("#chk_all").prop("checked", "");
		}
	});
}

// 判断选中的数量是否超过指定的上限
function moreThanMaxNum(count, num, message) {
	if(count > maxNum) {
		if(parent.$("#chk_all")[0].checked || num > maxNum) {
			parent.showWarningMessage(message);
			return true;
		}
	}
	return false;
}

