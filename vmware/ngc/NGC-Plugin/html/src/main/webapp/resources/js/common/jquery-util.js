
// Utility to serialize a form into Json data
// See http://benalman.com/projects/jquery-misc-plugins/#serializeobject
(function($,undefined){
   $.fn.serializeJson = function(){
     var obj = {};
     $.each( this.serializeArray(), function(i,o){
         var n = o.name,
         v = o.value;
         obj[n] = (obj[n] === undefined) ? v
           : $.isArray( obj[n] ) ? obj[n].concat( v )
           : [ obj[n], v ];
     });
     return JSON.stringify(obj);
   };
 })(jQuery);


// Set style of JQuery-ui accordion widget. See http://jqueryui.com/accordion/
$(function() {
   var accordion$ = $("#accordion");
   if (accordion$.length > 0) {
      $("#accordion").accordion({
        heightStyle: "content"
      });
   }
});

jQuery.debug = function(msg){
	if(window.console && window.console.log){
		window.console.log(msg);
	}
};

function isEmpObj(obj){
	return obj == 'undefined' || obj == null || obj == "";
}

function req(url,data){
	this.isAsy = true;
	this.type = "GET";
	this.url = url;
	this.data = data;
	this.dataType = "";
	//增大请求的超时时间
	this.timeout = 30 * 60 * 1000;
}

function handler(suc,err){
	this.doSuccess = suc;
	this.doError = err;
}

function sendMsg(req,handler){
    if(req.contentType == undefined || req.contentType == ""){
		contentType = 'application/json';
	}else{
		contentType = req.contentType;
	}
	if(req.type == "GET"){
	    req.data = encodeURI(req.data);
	}
	$.ajax({
		async:req.isAsy,
		type:req.type,
		url:encodeURI(req.url),
		data:req.data,
		dataType:req.dataType,
		contentType:contentType,
		timeout:req.timeout,
		success:function(resp){
			handler.doSuccess(resp);
		},
		error:function(){
			handler.doError();
		}
	});
}
