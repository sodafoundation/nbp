function dragTable_iframe(TableHeadID,iframeID,TableID){
    var tTD;
    var iframe01;
    var Table;
    var TableHead = document.getElementById(TableHeadID);
    	    for (j = 0; j < TableHead.rows[0].cells.length; j++) {
            TableHead.rows[0].cells[j].onmousedown = function () {

                tTD = this;
                if (event.offsetX > tTD.offsetWidth - 10) {
                    tTD.mouseDown = true;
                    tTD.oldX = event.x;
                    tTD.oldWidth = tTD.offsetWidth;
                }
            };
            TableHead.rows[0].cells[j].onmouseup = function () {

                if (tTD == undefined) tTD = this;
                tTD.mouseDown = false;
                tTD.style.cursor = 'default';
            };
            TableHead.rows[0].cells[j].onmouseleave= function (){
                 if(tTD.mouseDown == true){
                      if (tTD == undefined) tTD = this;
                        tTD.mouseDown = false;
                        tTD.style.cursor = 'default';
                }

            }
            TableHead.rows[0].cells[j].onmousemove = function () {

                if (event.offsetX > this.offsetWidth - 10)
                    this.style.cursor = 'col-resize';
                else
                    this.style.cursor = 'default';
                if (tTD == undefined) tTD = this;
                if (tTD.mouseDown != null && tTD.mouseDown == true) {
                    tTD.style.cursor = 'default';
                    if (tTD.oldWidth + (event.x - tTD.oldX) > 0)
                        tTD.width = tTD.oldWidth + (event.x - tTD.oldX);
                    var widthString=String(tTD.width);
                    widthString=widthString+"px";
                    tTD.style.width =widthString;
                    tTD.style.cursor = 'col-resize';
                    TableHead = tTD;
                    iframe01 = document.getElementById(iframeID);
                    Table =  iframe01.contentWindow.document.getElementById(TableID);
                    while (Table.tagName != 'TABLE'){
                    Table = Table.parentElement;
                    }
                    for (j = 0; j < Table.rows.length; j++) {
                        Table.rows[j].cells[tTD.cellIndex].width = tTD.width;
                    }
                }
            };
        }

}
var isDragClick = false;
function dragTable_table(TableHeadID,TableID){
    var tTD;
    var Table =  document.getElementById(TableID);
    var TableHead = document.getElementById(TableHeadID);
            TableHead.rows[0].onmouseup = function () {
                if (tTD != undefined) {
                    if(tTD.mouseDown == true){
                        tTD.mouseDown = false;
                        tTD.style.cursor = 'default';
                    }
                }
            };
            TableHead.rows[0].onmouseleave= function (){
                if (tTD != undefined) {
                    if(tTD.mouseDown == true){
                        tTD.mouseDown = false;
                        tTD.style.cursor = 'default';
                    }
                }
             }
    	    for (j = 0; j < TableHead.rows[0].cells.length; j++) {
            TableHead.rows[0].cells[j].onmousedown = function () {
                event.preventDefault();
                if (event.offsetX > this.offsetWidth - 10) {
                    tTD = this;
                    tTD.mouseDown = true;
                    isDragClick = true;
                    tTD.oldX = event.x;
                    tTD.oldWidth = tTD.offsetWidth;
                    tTD.oldLeft = tTD.left;
                }
            };

            TableHead.rows[0].cells[j].onmousemove = function () {
                event.preventDefault();
                if (event.offsetX > this.offsetWidth - 10)
                    this.style.cursor = 'col-resize';
                else
                    this.style.cursor = 'default';
                if (tTD == undefined) tTD = this;
                if (tTD.mouseDown != null && tTD.mouseDown == true) {
                    while (Table.tagName != 'TABLE') {
                    Table = Table.parentElement;
                    }
                    var dataRowCount = Table.rows.length;
                    if(dataRowCount > 0 ){
                        tTD.style.cursor = 'col-resize';
                        if (tTD.oldWidth + (event.x - tTD.oldX) > 0)
                            tTD.width = tTD.oldWidth + (event.x - tTD.oldX);
                        var widthString=String(tTD.width);
                        widthString=widthString+"px";
                        tTD.style.width=widthString;
                        for (j = 0; j < dataRowCount; j++) {
                            Table.rows[j].cells[tTD.cellIndex].width = tTD.width;
                        }
                    }else{
                        tTD.style.cursor = 'default';
                        tTD.mouseDown = false;
                    }
                }
            };
            TableHead.rows[0].cells[j].onclick= function (){
                if (!isDragClick) {
                    _sortTable(Table,this.cellIndex);
                }else{
                    isDragClick = false;
                }
            };
        }

}

function _sortTable(table,Idx){
    var tbody = table.tBodies[0];
    var tr = tbody.rows;

    var rowCount = tr.length;
    var trValue = new Array();
    for (var i=0; i<rowCount; i++ ) {
    	trValue[i] = tr[i]; 
    }

    if (tbody.sortCol == Idx) {
    	trValue.reverse();
    } else {
    	trValue.sort(function(tr1, tr2){
    		var value1 = tr1.cells[Idx].innerHTML;
    		var value2 = tr2.cells[Idx].innerHTML;
    		return value1.localeCompare(value2);
    	});
    }

    var fragment = document.createDocumentFragment();  
    for (var i=0; i<rowCount; i++ ) {
    	fragment.appendChild(trValue[i]);
    }
    tbody.appendChild(fragment);
    tbody.sortCol = Idx;
}