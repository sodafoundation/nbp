var ns = org_opensds_storage_devices;
var moref = parent.moref;
var serverGuid = parent.serverGuid;
var storaePoolCapChart;  // chart for storagePool
var interval = 30000; // set update time 30000s

$(document).ready(function() {
    var volume = {
        name: "Bill",
        capacity : "10G",
        storagePoolName : "testPool",
        storagePoolUseedCapacity : "60G",
        storagePoolUsage: 0.67
    };
    $("#volumeTb").append("<tr><td style='width:65%'><b>Volume Name</b></td><td>" + volume.name + "</td>" +
                            "<tr><td style='width:65%'><b>Volume Total Capacity</b></td><td>" + volume.capacity + "</td>" +
                            "<tr><td style='width:65%'><b>Stroge Name</b></td><td>" + volume.storagePoolName + "</td>" +
                            "<tr><td style='width:65%'><b>Stroge Total Capacity</b></td><td>" + volume.storagePoolUseedCapacity + "</td>" +
                            "<tr><td style='width:65%'><b>Stroge Pool Capactiy Usage</b></td><td>" + volume.storagePoolUsage * 100 + "%</td>");
    var option = {
        tooltip : {
            formatter: "{a} <br/>{b} : {c}%"
        },
        toolbox: {
            feature: {
                restore: {},
                saveAsImage: {}
            }
        },
        series: [
            {
                name: 'Storage Pool Capaptiy',
                type: 'gauge',
                detail: {formatter:'{value}%'},
                data: [{value: 60, name: 'Pool Capacity usage'}]
            }
        ]
    };

    option.series[0].data[0].value = volume.storagePoolUsage * 100;
    var childDiv = $("<div></div>");
    var childId = "volume" + "1";
    childDiv.attr("id", childId);
    childDiv.attr("style", "width: 400px; height:375px; margin:0 auto;");
    $("#storagePoolCharts").append(childDiv);
    storaePoolCapChart = echarts.init(document.getElementById(childId));
    storaePoolCapChart.setOption(option, true);
    /*
    var queryUrl = ns.webContextPath + "/rest/data/datastore/getInfo/" + moref + "?serverGuid="
        + serverGuid + "&t=" + new Date();
    $.getJSON(encodeURI(queryUrl), function (resp) {
        if (resp.msg != null){
            var arr = eval(resp.data);
            for ( var i = 0; i < arr.length; i++) {
                var jsonObj = arr[i];
                $("#volumeTb").
                option.series[0].data[0].value = jsonObj.storagePoolUsage;
                storaePoolCapChart = echarts.init(document.getElementById('storagePoolChart'));
                storaePoolCapChart.setOption(option, true);
            }

        }
    });*/
});