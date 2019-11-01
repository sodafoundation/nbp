var pageParam = {
	"pager": {
		"sizeWidth": "storage.plugin.action.base.pager.page.size.width",
		"total": "of",
		"last": "Last Page",
		"next": "Next Page",
		"display1": "Items",
		"perPage": "",
		"displayWidth": "storage.plugin.action.base.pager.page.display.width",
		"ofDataPerPage": "100 items per page",
		"prev": "Previous Page",
		"ofData": "",
		"jump": "GO",
		"display2": "to",
		"display3": "Total:",
		"page": "",
		"prefix": "Page",
		"first": "First Page"
	},
	"chk": {
		"child": "chkchild",
		"all": "chk_all"
	},
	"button": {
		"cancel": "Cancel",
		"ok": "OK",
		"close": "Close"
	},
	"regist": {
		"notconnected": "Registration failed. Failed to connect to the vCenter.",
		"localunsupport": "Registration failed. The language is not supported."
	},
	"validate": {
		"usernameLengthMax": "username cant longer than 32",
		"minmax2": "to",
		"passwordLengthMax": "password cant longer than 32",
		"passwordLength": "Password length must over 8.",
		"minmax1": "The [xxx] must be an integer ranging from",
		"special2": "The [xxx] can contain only digits, letters, and special characters.",
		"special": "The [xxx] can contain only digits, letters, underscores (_), period (.), and hyphens (-), and must start with a letter or underscore (_). The [xxx] cannot be empty.",
		"ipFormat": "[xxx] is a reserved IP address. Please retry.",
		"userInfoverify": "username or password is not correct",
		"ipRange": "The start IP address cannot be greater than the end IP address.",
		"required": "The [xxx] cannot be empty.",
		"ip": "[xxx] must be in the format of 192.168.100.1.\t\t\t\t\t\t\t\t\t\t\t\t\t"
	}
};

(function($) {
	var bigPage = new

	function() {
		this.cssWidgets = [];
		this.ajaxpage = function(param) {
			this.config = {

				container: $("#pager1"),

				data: null,

				ajaxData: {
					url: "",
					params: {}
				},

				pageSize: 100,

				toPage: 1,

				cssWidgetIds: [],

				position: "down",

				totalPages: 1,

				totalRows: 10,

				maxPageNumCount: 10,

				cssNext: '.next',
				cssPrev: '.prev',
				cssFirst: '.first',
				cssLast: '.last',
				cssJump: '.ytb-sep',

				callback: null
			};
			$.extend(this.config, param);

			
			this.isFirstPage = function() {
				if(this.config.toPage == 1) {
					return true;
				}
				return false;
			};

		
			this.firstPage = function() {
				if(this.config.toPage == 1) {
					return this;
				}
				var iframeid = this.config.ajaxData.params.iframeId;
				$("#" + iframeid).attr("src", "");
				var loading = this.config.ajaxData.params.loaddingId;
				$("#" + loading).css("display", "block");
				this.config.toPage = 1;
				this.applyBuildTable();
				return this;
			};


			this.prevPage = function() {
				if(this.config.toPage <= 1) {
					return this;
				}
				var iframeid = this.config.ajaxData.params.iframeId;
				$("#" + iframeid).attr("src", "");
				var loading = this.config.ajaxData.params.loaddingId;
				$("#" + loading).css("display", "block");
				this.config.toPage--;
				this.applyBuildTable();
				return this;
			};

			this.nextPage = function() {
				if(this.config.toPage >= this.config.totalPages) {
					return this;
				}
				var iframeid = this.config.ajaxData.params.iframeId;
				$("#" + iframeid).attr("src", "");
				var loading = this.config.ajaxData.params.loaddingId;
				$("#" + loading).css("display", "block");
				this.config.toPage++;
				this.applyBuildTable();
				return this;
			};

			this.lastPage = function() {
				if(this.config.toPage == this.config.totalPages) {
					return this;
				}
				var iframeid = this.config.ajaxData.params.iframeId;
				$("#" + iframeid).attr("src", "");
				var loading = this.config.ajaxData.params.loaddingId;
				$("#" + loading).css("display", "block");
				this.config.toPage = this.config.totalPages;
				this.applyBuildTable();
				return this;
			};

			
			this.isLastPage = function() {
				if(this.config.toPage == this.config.totalPages) {
					return true;
				}
				return false;
			};

		
			this.skipPage = function(toPage_) {
				var numberValue = Number(toPage_);
				var totalPage = this.config.totalPages;
				if($.trim(numberValue) == "") {
					toPage = this.config.toPage;
				} else if(isNaN(numberValue)) {
					toPage = this.config.toPage;
				} else {
					with(this.config) {
						toPage = numberValue;
						if(toPage < 1 || toPage > totalPage) {
							toPage = toPage < 1 ? 1 : totalPage;
						}
					}
				}
				var iframeid = this.config.ajaxData.params.iframeId;
				$("#" + iframeid).attr("src", "");
				var loading = this.config.ajaxData.params.loaddingId;
				$("#" + loading).css("display", "block");
				this.applyBuildTable();
				return this;
			};

		
			this.getSubData = function() {
				if(this.config.data != null && $.isArray(this.config.data)) {
					var totalItems = this.config.totalItems;
					if(totalItems <= 0) {
						return [];
					}
					var startRow = (this.config.toPage - 1) * this.config.pageSize;
					var endRow = this.config.toPage * this.config.pageSize;
					if(startRow > totalItems) {
						return [];
					}
					if(endRow > totalItems) {
						endRow = totalItems;
					}
					return this.config.data.slice(startRow, endRow);
				} else if(this.config.ajaxData.data && $.isArray(this.config.ajaxData.data)) {
					return this.config.ajaxData.data;
				} else {
					return [];
				}
			};

			this.search = function(searchParam) {
				this.config.ajaxData.params = this.config.ajaxData.params || {};
				$.extend(this.config.ajaxData.params, searchParam);
				this.config.toPage = 1;
				this.applyBuildTable();
			};

			this.applyBuildTable = function() {
				var $table = this;

				var page = this.config.container;
				var data = this.config.data;
				if(data != null && $.isArray(data)) {
					this.config.totalItems = data.length;
					this.config.totalPages = totalPageFun(data.length, this.config.pageSize);
					buildTable();
				} else if(!bigPage.isNull(this.config.ajaxData.url)) {
					this.config.ajaxData.params = this.config.ajaxData.params || {};
					$.extend(this.config.ajaxData.params, {
						toPage: this.config.toPage,
						pageSize: this.config.pageSize
					});


					var poolRequest = $table.config.ajaxData.params.iframeId;
					if(poolRequest == "storagepoolListFrame") {
						$("#storagelunListFrame").attr("src", "");
						$("#pager2").remove();
						setFilter_lun("NAME", default_input);
						setPoolId("");
					}
					if(this.config.totalItems == undefined) {

						$.ajax({
							async: true,
							type: "POST",
							url: encodeURI(this.config.ajaxData.url),
							data: this.config.ajaxData.params,
							dataType: "json",
							timeout: 30 * 60 * 1000,
							success: function(resp) {

								var iframeId = $table.config.ajaxData.params.iframeId;
								if(iframeId == "snapshotFrame" || iframeId == "tabFrame_lun" || iframeId == "unmappedlunTabFrame" || iframeId == "mappedlunTabFrame") {
									if($table.config.ajaxData.params.data_params != loadpage2_data_params) {
										return;
									}
								}

	
								if(null != resp.errorCode) {
									var loading = $table.config.ajaxData.params.loaddingId;
									$("#" + loading).css("display", "none");
									var errorloading = $table.config.ajaxData.params.errorloaddingId;
									$("#" + errorloading).css("display", "block");
									var description = "<span style='width: 0; height: 100%; display: inline-block; vertical-align: middle;'></span>" + resp.errorDesc;
									$("#" + errorloading).html(description);
									return;
								}
	
								$table.config.totalItems = resp.data;
								if($table.config.totalItems != 0) {
									$table.config.totalPages = totalPageFun(resp.data, $table.config.pageSize);

									var iframeid = $table.config.ajaxData.params.iframeId;
									var data_url = $table.config.ajaxData.params.data_url;
									var data_params = $table.config.ajaxData.params.data_params;
									var startItems = ($table.config.toPage - 1) * $table.config.pageSize;
									$("#" + iframeid).attr("src", encodeURI(data_url + "?start=" + startItems + "&pagesize=" + $table.config.pageSize + data_params));
									buildTable("frameid: " + $("#" + iframeid).attr("src"));
								} else {
									var loading = $table.config.ajaxData.params.loaddingId;
									$("#" + loading).css("display", "none");
									executeCallback();
								}

							},
							error: function() {
								var loading = $table.config.ajaxData.params.loaddingId;
								$("#" + loading).css("display", "none");
							},
							beforeSend: function() {},
							complete: function() {}
						});
					}else{
						if($table.config.totalItems != 0) {
							$table.config.totalPages = totalPageFun($table.config.totalItems, $table.config.pageSize);

							var iframeid = $table.config.ajaxData.params.iframeId;
							var data_url = $table.config.ajaxData.params.data_url;
							var data_params = $table.config.ajaxData.params.data_params;
							var startItems = ($table.config.toPage - 1) * $table.config.pageSize;
							$("#" + iframeid).attr("src", encodeURI(data_url + "?start=" + startItems + "&pagesize=" + $table.config.pageSize + data_params));
							buildTable("frameid: " + $("#" + iframeid).attr("src"));
						} else {
							var loading = $table.config.ajaxData.params.loaddingId;
							$("#" + loading).css("display", "none");
							executeCallback();
						}
					}
				}

				function totalPageFun(totalItems, pageSize) {
					if(totalItems <= 0) return 0;
					var totalPage = Math.ceil(totalItems / pageSize);
					return isNaN(totalPage) ? 0 : totalPage;
				};

				function buildTable() {
					bigPage.applyCssWidget($table);
					if($table.config.callback && $.isFunction($table.config.callback)) {
						$table.config.callback($table);
					}
				}

				function executeCallback() {
					if($table.config.callback && $.isFunction($table.config.callback)) {
						$table.config.callback($table);
					}
				}
			};

			this.applyBuildTable();
			return this;
		};

		this.isNull = function(obj) {
			if(obj == null || $.trim(obj) == "" || typeof(obj) == "undefined") {
				return true;
			}
			return false;
		};

		this.addCssWidget = function(cssWidget) {
			this.cssWidgets.pushEx(cssWidget);
			return this;
		};

		this.applyCssWidget = function($table) {
			var this_ = this;
			var cssWidgetIds = $table.config.cssWidgetIds;
			if(cssWidgetIds.length <= 0) {
				cssWidgetIds[0] = "ajaxpageBar1";
			} else {
				var hasAppendToTable = false;
				for(var i = 0; i < cssWidgetIds.length; i++) {
					if(cssWidgetIds[i] == "appendToTable") {
						hasAppendToTable = true;
					}
				}
				if(!hasAppendToTable) {
					cssWidgetIds = ["appendToTable"].concat(cssWidgetIds);
				}
			}

			for(var i = 0; i < cssWidgetIds.length; i++) {
				var cssWidget = getCssWidgetById(cssWidgetIds[i]);
				if(cssWidget) {
					cssWidget.format($table);
				}
			}

			function getCssWidgetById(name) {
				if(this_.isNull(name)) {
					return false;
				}
				var len = this_.cssWidgets.length;
				for(var i = 0; i < len; i++) {
					if(this_.cssWidgets[i].id.toLowerCase() == name.toLowerCase()) {
						return this_.cssWidgets[i];
					}
				}
				return false;
			}
		};

		Array.prototype.pushEx = function(obj) {
			var a = true;
			for(var i = 0; i < this.length; i++) {
				if(this[i].id.toLowerCase() == obj.id.toLowerCase()) {
					this[i] = obj;
					a = false;
					break;
				}
			}
			if(a) {
				this.push(obj);
			}
			return this.length;
		};

	};

	$.extend({
		bigPage: bigPage
	});
	$.fn.bigPage = bigPage.ajaxpage;

	$.bigPage.addCssWidget({
		id: "appendToTable",
		format: function($table) {
			var subData = $table.getSubData();
			var $tBody = $table.find("tbody:first");
			var trsArray = [];
			for(var i = 0; i < subData.length; i++) {
				var cellVaues = subData[i];
				var trArray = [];
				trArray.push("<tr>");
				for(var j = 0; j < cellVaues.length; j++) {
					trArray.push("<td>");
					trArray.push(cellVaues[j]);
					trArray.push("</td>");
				}
				trArray.push("</tr>");
				trsArray.push(trArray.join(""));
			}
			$tBody.html(trsArray.join(""));
		}
	});

	function moveToFirstPage(table) {
		var c = table.config;
		c.page = 0;
		moveToPage(table);
	}

	$.bigPage.addCssWidget({
		id: "ajaxpageBar1",
		format: function($table) {

			var displayNum = $table.config.toPage * $table.config.pageSize;
			if(displayNum >= $table.config.totalItems) {
				displayNum = $table.config.totalItems;
			}

			var footPageHtml = '<div id="' + $table.config.container + '" class="pager"><div class="pagerDiv1"><table width="400px" cellpadding="0" cellspacing="0"><tbody>' + '<tr><td align="center" style="vertical-align: middle" width="28px"><img src="' + ns.webContextPath + '/assets/images/first.png" class="first"></img></td>' + '<td align="left" style="vertical-align: middle" width="18px"><img src="' + ns.webContextPath + '/assets/images/prev.png" class="prev"/></img></td>' + '<td align="left" style="vertical-align: middle" width="13px"><span class="ytb-sep"></span></td>' + '<td align="left" style="vertical-align: middle" width="' + pageParam.pager.displayWidth + '"><span id="pageDisplay" class="pageDisplay">'

				+
				'<table cellpadding="0" cellspacing="0" width="70%"><tr><td>' + pageParam.pager.prefix + '</td><td><input type="text" width="15px" id="txtPageNum_' + $table.config.container + '" ' + 'style="background-image: url(' + ns.webContextPath + '/assets/images/icon_input.png);height: 15px;padding-left: 5px;width: 25px;border: 1px solid #E5E5E5;"' + 'value="' + $table.config.toPage + '"/>' + '</td><td>' + pageParam.pager.page + '</td><td>&nbsp;&nbsp;&nbsp;</td><td>' + pageParam.pager.total + '&nbsp;' + $table.config.totalPages + "&nbsp;" + pageParam.pager.page + "</td></tr></table>"

				+
				'</span></td>' + '<td align="left" style="vertical-align: middle" width="13px"><span class="ytb-sep"></span></td>' + '<td align="left" style="vertical-align: middle" width="' + pageParam.pager.sizeWidth + '"><span id="pageSize" class="pageDisplay">'

				+
				'<table cellpadding="0" cellspacing="0" width="100%"><tr><td>' + $table.config.pageSize + " items per page" /*pageParam.pager.ofDataPerPage*/ + '</td><td>' + "<input type='text' id='txtPageSize' style='display: none' value='" + $table.config.pageSize + "'/>" + "</td></tr></table>"

				+
				'</span></td>' + '<td align="left" style="vertical-align: middle" width="13px"><span class="ytb-sep"></span></td>' + '<td align="left" style="vertical-align: middle" width="20px"><img src="' + ns.webContextPath + '/assets/images/next.png" class="next"></img></td>' + '<td align="left" style="vertical-align: middle" width="16px"><img src="' + ns.webContextPath + '/assets/images/last.png" class="last"></img></td>' + '<td align="left" style="vertical-align: middle" width="13px"><span class="ytb-sep"></span></td>' + '<td width="30px" style="vertical-align: middle"><span id="pagerGO_' + $table.config.container + '" class="ytb-go">' + pageParam.pager.jump + '</span></td>' + '<td>&nbsp;</td></tr>' + '</tbody></table></div>'


				+
				'<div class="pageDiv2"><table width="251px" cellpadding="0" cellspacing="0">' + '<tr><td width="100%" align="right" style="vertical-align: middle"><span id="display" class="pageDisplaySpan">'

				+
				pageParam.pager.display1 + "&nbsp;" + (($table.config.toPage - 1) * $table.config.pageSize + 1) + "&nbsp;" + pageParam.pager.display2 + "&nbsp;" + displayNum + "&nbsp;" + pageParam.pager.display3 + "&nbsp;" + $table.config.totalItems + "&nbsp;" + pageParam.pager.ofData

				+
				'</span>' + '</td></tr></table></div>';
			$("#pager-" + $table.config.ajaxData.params.iframeId).empty();
			if($table.config.position == "up") {
				$table.before(footPageHtml);
			} else if($table.config.position == "both") {
				$table.before(footPageHtml);
				$table.after(footPageHtml);
			} else {
				var iframeid = $table.config.ajaxData.params.iframeId;
				$("#pager-" + iframeid).append(footPageHtml);
			}

			$footDiv = $table.siblings("div[id='" + $table.config.container + "']");

			$('#' + $table.config.container + ' .first').unbind('click'); 
			$('#' + $table.config.container + ' .first').click(function() {
				$table.firstPage();
				return false;
			}).mousemove(function() {
				$(this).css("cursor", "pointer");
			});

			$('#' + $table.config.container + ' .next').unbind('click'); 
			$('#' + $table.config.container + ' .next').click(function() {
				$table.nextPage();
				return false;
			}).mousemove(function() {
				$(this).css("cursor", "pointer");
			});

			$('#' + $table.config.container + ' .prev').unbind('click'); 
			$('#' + $table.config.container + ' .prev').click(function() {
				$table.prevPage();
				return false;
			}).mousemove(function() {
				$(this).css("cursor", "pointer");
			});

			$('#' + $table.config.container + ' .last').unbind('click'); 
			$('#' + $table.config.container + ' .last').click(function() {
				$table.lastPage();
				return false;
			}).mousemove(function() {
				$(this).css("cursor", "pointer");
			});

			$("#pagerGO_" + $table.config.container).bind("click",
				function(event) {
					$table.skipPage(parseInt($("#txtPageNum_" + $table.config.container).val()));
				}).mousemove(function() {
				$(this).css("cursor", "pointer");
			});;

			$("#txtPageNum_" + $table.config.container, $footDiv).unbind('keyup'); 
			$("#txtPageNum_" + $table.config.container).bind("keyup",
				function(event) {
					var o = $(this);
					var total = $table.config.totalPages;
					if($.trim(o.val()) == "") {
						o.val("");
					} else if(isNaN(o.val())) {
						o.val("");
					} else if(o.val() > total) {
						o.val(total);
					} else if(o.val() < 1) {
						o.val(1);
					}
					if(event.keyCode == 13) {
						$table.skipPage(parseInt($("#txtPageNum_" + $table.config.container).val()));
					}
				}).bind("blur",
				function() {
					var o = $(this);
					var pageNum = parseInt($table.config.toPage);
					if(isNaN(o.val()) || $.trim(o.val()) == "") {
						o.val(pageNum);
					}
				});
		}
	});

})(jQuery);