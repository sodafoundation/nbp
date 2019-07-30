/* 下面是使表格能够全选的代码 **************************/
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
		if($(this).prop("checked")){// 全选
			$(tbSel).prop("checked",true);
			tabModel.selCount = tabModel.rowCount;
		}
		else{// 全不选
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
			if(tabModel.rowCount == tabModel.selCount)//不全选-->全选
			{
				$(thSel).prop("checked",true);
			}
		}
		else{
			$.debug($(this).prop("id")+"取消....");
			tabModel.selCount--;
			if(tabModel.rowCount > tabModel.selCount)//全选-->不全选
			{
				$(thSel).prop("checked",false);
			}
		};
		$.debug("click 后的selCount: "+tabModel.selCount);
	});
}
/* 下面是ROM转换  ************************/
/**
 * @param objs 是要填充的对象数组,请在外部做非空校验
 * @param cloId 要克隆的行id选择器
 * @param tbsel 表格体的选择器
 */
function a2t(tbsel,cloId,objs){
	for ( var i = 0; i < objs.length; i++) {
		var $lunRow = $(cloId).clone(true);// 克隆行
		$lunRow.attr("id", cloId + i);
		obj2rel($lunRow, objs[i]);// 赋值
		$(tbsel).append($lunRow);// 追加行
		$lunRow.show();// 显示行
	}
}
/** 将obj属性值填充给tr,对于属性顺序和td顺序无要求,但是是通过name属性值来匹配bean属性 **/
function obj2rel($tr, obj) {
	$tr.children("td").each(function(num, td) {
		if(!isEmpObj($(td).attr("name"))){
		    // chg 20181101 : ie11 compatibility bug
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
/** table记录转为对象数组 */
function t2a(tableId) {
	var objList = new Array();
	$(tableId).find("tbody tr").each(function(nun, tr) {
		var length = $(tr).find(":checked").length;// 判断是否有选中行
		if (length > 0) {
			var obj = r2o(tr);
			objList.push(obj);
		}
	});
	return objList;
}
/** row转为obj */
function r2o(tr) {
	var obj = new Object();
	$(tr).children("td").each(
			function(num, td) {// 将每一行
				if (!isEmpObj($(td).attr("name"))) {
					obj[$(td).attr("name")] = td.innerHTML;
				}
			});
	return obj;
}