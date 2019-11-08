
$(document).ready(function() {
	initData();
	bindEvent();
});

function initData() {
	var request = new Object();
	request = GetRequest();

	var url = "";
	var lunReq = "";
	if(parent.lun_fs_flag == "lun") {
		url = parent.ns.webContextPath + "/rest/data/host/snapshot";
		lunReq = new req(url, "storageId=" + parent.devObj.id + "&volumeId=" + parent.lunObj.id + "&start=" + request["start"] + "&count=" + request["pagesize"] + "&t=" + new Date());
	} else if(parent.lun_fs_flag == "fs") {
		url = parent.ns.webContextPath + "/rest/nfsdata/fsSnapshot";
		lunReq = new req(url, "storageId=" + parent.devObj.id + "&fsId=" + parent.fsObj.id + "&start=" + request["start"] + "&count=" + request["pagesize"] + "&t=" + new Date());
	}
	var lunhandler = new handler(function doSuccess(resp) {
		if(resp.errorCode) {
			parent.$("#divLoadingSnapshot").hide();
			$("#divError").text(resp.errorDesc).show();
			return;
		}
		resp.data.forEach(function(item, index, array) {
			item.activatedAt = new Date(item.activatedAt);
		})
		a2t("#snapshotTable tbody", "#cloneSnap", resp.data);
		$("#snapshotTable td[name='activatedAt']").each(function() {
			$(this).html(new Date($(this).html()))
		});
		$("#cloneSnap").remove();
		scroll("snapshotTab", "snapshotTabDiv", 1, parent.divhead_id_snapshot, "snapshotTable");
		parent.init();
		parent.$("#divLoadingSnapshot").hide();
	}, function doFailed() {
		parent.$("#divLoadingSnapshot").hide();
	});
	sendMsg(lunReq, lunhandler);
}

function bindEvent() {

	$("#snapshotTable tbody tr").each(function() {
		$(this).find("td").each(function() {
			if($(this).index() != 2) {
				$(this).click(function() {
					console.log("click the tr");
					if($(this).parent("tr").find("td").eq(2).find("input").is(":checked")) {
						$(this).parent("tr").find("td").eq(2).find("input").prop("checked", false);
					} else {
						$(this).parent("tr").find("td").eq(2).find("input").prop("checked", true);
					}
					clickSingleCheckBox();
				});
			}
		});
	});
}

function clickSingleCheckBox() {

	var num = 0;
	var $singleChkbox = parent.$("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']");
	$singleChkbox.each(function(i) {
		if(this.checked) {
			num++;
		}
	});

	if(num == parent.rowNum) {
		if(!parent.$("#chk_all")[0].checked) {
			parent.$("#chk_all").prop("checked", "checked");
		}
	} else {
		parent.$("#chk_all").prop("checked", "");
	}
	if(num > 0) {
		parent.$("#delSnapBtn").prop("disabled", "");
		parent.$("#delSnapBtn").removeClass("disabled");
		parent.$("#delSnapBtn .plugin_button_div").css("cursor", "pointer");
		if(num == 1) {
			if(parent.lun_fs_flag == "lun") {
				var lunRunningStatus = parent.$("#snapshotFrame")[0].contentWindow.$("input[id^='snapCheckbox_']:checked").parent().parent().find("td[name='runningStatus']").text();
				if(lunRunningStatus == "INACTIVATED") {
					parent.$("#recoverBtn").prop("disabled", "disabled");
					parent.$("#recoverBtn").addClass("disabled");
					parent.$("#recoverBtn .plugin_button_div").css("cursor", "default");
				} else {
					parent.$("#recoverBtn").prop("disabled", "");
					parent.$("#recoverBtn").removeClass("disabled");
					parent.$("#recoverBtn .plugin_button_div").css("cursor", "pointer");
				}
			} else if(parent.lun_fs_flag == "fs") {
				parent.$("#recoverBtn").prop("disabled", "");
				parent.$("#recoverBtn").removeClass("disabled");
				parent.$("#recoverBtn .plugin_button_div").css("cursor", "pointer");
			}
		} else {
			parent.$("#recoverBtn").prop("disabled", "disabled");
			parent.$("#recoverBtn").addClass("disabled");
			parent.$("#recoverBtn .plugin_button_div").css("cursor", "default");
		}
	} else {
		parent.$("#delSnapBtn").prop("disabled", "disabled");
		parent.$("#recoverBtn").prop("disabled", "disabled");
		parent.$("#delSnapBtn").addClass("disabled");
		parent.$("#recoverBtn").addClass("disabled");
		parent.$("#delSnapBtn .plugin_button_div").css("cursor", "default");
		parent.$("#recoverBtn .plugin_button_div").css("cursor", "default");
		parent.$("#chk_all").prop("checked", "");
	}

}

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
	var url = location.search;   
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

function trim(str) {
	return str.replace(/(^\s*)|(\s*$)/g, "");
}