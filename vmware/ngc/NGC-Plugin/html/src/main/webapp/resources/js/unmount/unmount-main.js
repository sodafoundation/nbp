function refreshData() {
	loadpage2_data_params = "&filterType=" + filterType
			+ "&filterValue=" + filterValue + "&serverGuid=" + serverGuid + "&t="
			+ new Date();
	$("#chk_all").prop("checked", false);
	$("#chk_all").attr("disabled", "disabled");
	$("#btnUnmount").prop("disabled", "disabled");
	$("#diverrorLUN").hide();
	if ($("#" + divhead_id).length > 0) {
		$("#" + divhead_id).width($("#divMain").width() - 22);
	}
	$('#mappedlunTabFrame').prop("src", "");
	$("#pager1").remove();
	$("#divLoadingMappedLUN").css("display", "block");
	var url = ns.webContextPath + "/rest/data/host/unmountableVolumeList/count/" + hostId
			+ "?filterType=" + filterType
			+ "&filterValue=" + filterValue + "&serverGuid=" + serverGuid + "&t="
			+ new Date();
    $("#mappedLunList").bigPage(
			{
				container : "pager1",
				ajaxData : {
					url : encodeURI(url),
					params : {
						loaddingId : "divLoadingMappedLUN",
						errorloaddingId : "diverrorLUN",
						iframeId : "mappedlunTabFrame",
						data_url : ns.webContextPath + "/resources/html/unmount/lunTab.html",
						data_params : loadpage2_data_params
					}
				},
				pageSize : pagesize_lun,
				toPage : toPage_lun,
				position : "down",
				callback : null
			});
}

//调整页面结构
function changesize() {
	var divMainHeight = $("#divMain").height();

	var topHeight = $("#top").height();
	var lineTop = 5 + topHeight + 5;
	var buttonsTop = lineTop + 2 + 5;
	//分割线的位置
	$("#line").css("top", lineTop);
	//按钮的位置
	$("#buttons").css("top", buttonsTop);
	var buttonsHeight = $("#buttons").height();
	//表格的位置 和高度
	var tableTop = buttonsTop + buttonsHeight + 5;
	var tableHeight = divMainHeight - tableTop - 3;
	$("#mappedLunListDiv").height(tableHeight - 63);
	
	$("#mappedlunTabFrame").height($("#mappedLunListDiv").height() - $("#mappedLunList").height() - 24);
}
