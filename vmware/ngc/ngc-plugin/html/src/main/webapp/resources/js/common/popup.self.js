       var pop =null;
       function ShowIframe(id,title,contentUrl,width,height,callback)
        {
           pop=new Popup({ contentType:1,isReloadOnClose:false,width:width,height:height,isSupportDraging:true,popId:id});
           pop.setContent("contentUrl",contentUrl);
           pop.setContent("title",title);
           pop.setContent("closeCallBack",callback);
           pop.build();
           pop.show();
        }
        function ShowHtmlString(title,strHtml,width,height)
        {
            pop=new Popup({ contentType:2,isReloadOnClose:false,width:width,height:height,isSupportDraging:true});
            pop.setContent("contentHtml",strHtml);
            pop.setContent("title",title);
            pop.build();
            pop.show();
       }
        function ShowConfirm(title,imagePath,confirmCon,width,height,callback)
        {
            pop=new Popup({ contentType:3,isReloadOnClose:false,width:width,height:height,isSupportDraging:true});
            pop.setContent("title",title);
            pop.setContent("confirmCon",confirmCon);
            pop.setContent("callBack",callback);
            pop.setContent("imagePath",imagePath);
            pop.build();
            pop.show();
            pop.focus();
            return pop;
        }
        function ShowConfirmDelResources(title,confirmCon,width,height,callback)
        {
            pop=new Popup({ contentType:6,isReloadOnClose:false,width:width,height:height,isSupportDraging:true});
            pop.setContent("title",title);
            pop.setContent("confirmCon",confirmCon);
            pop.setContent("callBack",callback);
            pop.build();
            pop.show();
            return pop;
        }
        function ShowAlert(title,imagePath,alertCon,width,height)
        {
            pop=new Popup({ contentType:4,isReloadOnClose:false,width:width,height:height});
            pop.setContent("title",title);
            pop.setContent("imagePath",imagePath);
            pop.setContent("alertCon",alertCon);
            pop.build();
            pop.show();
            pop.focus();
        }
        function ShowAlertCallBack(id,title,imagePath,alertCon,width,height,callback)
        {
        	pop=new Popup({ contentType:5,isReloadOnClose:false,width:width,height:height,popId:id});
        	pop.setContent("title",title);
        	pop.setContent("imagePath",imagePath);
        	pop.setContent("alertCon",alertCon);
        	pop.setContent("callBack",callback);
        	pop.setContent("closeCallBack",callback);
        	pop.build();
        	pop.show();
        	pop.focus();
        	return pop;
        }
        function ShowAlert2(title,alertCon,width,height,callback)
        {
        	pop=new Popup({ contentType:5,isReloadOnClose:false,width:width,height:height});
        	pop.setContent("title",title);
        	pop.setContent("alertCon",alertCon);
        	pop.setContent("callBack",callback);
        	pop.build();
        	pop.show();
        }

        function ShowCallBack(para)
        {
            var o_pop = para["obj"]
            var obj = document.getElementById(para["id"]);
            o_pop.close();
            obj.click();
        }
        function ClosePop()
        {
	        pop.close();
        }
