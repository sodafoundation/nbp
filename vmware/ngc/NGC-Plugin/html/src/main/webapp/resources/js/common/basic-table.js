function TabModel() {
	this.thSel = "";
	this.tbSel = "";
	this.rowCount = 0;
	this.selCount = 0;
}
function eSelAll(tabModel){
	initTabModel(tabModel);
	bindEvent(tabModel);
}
function initTabModel(tabModel){
	tabModel.rowCount = $(tabModel.tbSel+" tr ").length;
	$.debug("initData rowCount: "+tabModel.rowCount);
	tabModel.selCount = $(tabModel.tbSel+" :checked ").length;
	$.debug("initData selCount: "+tabModel.selCount);
}
function bindEvent(tabModel){
	$(tabModel.thSel).click(function(){
		$.debug("selAll click 前的selCount: "+tabModel.selCount);
		if($(this).prop("checked")){
			$(tbSel).prop("checked",true);
			tabModel.selCount = tabModel.rowCount;
		}
		else{
			$(tbSel).prop("checked",false);
			tabModel.selCount = 0;
		}
		$.debug("disAll click 后的selCount: "+tabModel.selCount);
	});
	$(tabModel.tbSel).click(function(){
		$.debug("click 前的selCount: "+tabModel.selCount);
		if($(this).prop("checked")){
			$.debug($(this).prop("id")+"选中....");
			tabModel.selCount++;
			if(tabModel.rowCount == tabModel.selCount)
			{
				$(thSel).prop("checked",true);
			}
		}
		else{
			$.debug($(this).prop("id")+"取消....");
			tabModel.selCount--;
			if(tabModel.rowCount > tabModel.selCount)
			{
				$(thSel).prop("checked",false);
			}
		};
		$.debug("click 后的selCount: "+tabModel.selCount);
	});
}

function a2t(tbsel,cloId,objs){
	for ( var i = 0; i < objs.length; i++) {
		var $lunRow = $(cloId).clone(true);
		$lunRow.attr("id", cloId + i);
		obj2rel($lunRow, objs[i]);
		$(tbsel).append($lunRow);
		$lunRow.show();
	}
}

function obj2rel($tr, obj) {
	$tr.children("td").each(function(num, td) {
		if(!isEmpObj($(td).attr("name"))){
		    //  ie11 compatibility bug
		    if(isEmpObj( obj[$(td).attr("name")]) && obj[$(td).attr("name")] != false)
            {
            	td.innerHTML = "";
            	td.title = "";
            }
            else
            {
            	td.innerHTML = obj[$(td).attr("name")];
            	td.title = td.innerHTML;
            }
		}
	});
}

function t2a(tableId) {
	var objList = new Array();
	$(tableId).find("tbody tr").each(function(nun, tr) {
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
			function(num, td) {
				if (!isEmpObj($(td).attr("name"))) {
					obj[$(td).attr("name")] = td.innerHTML;
				}
			});
	return obj;
}