// --------------------------------------------
//
// --------------------------------------------

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	//	"container/list"
)

/*---------------------
 */
func HandlerSPLINE(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/spline" {
		http.NotFound(w, r)
		return
	}

	var pp PieData
	pp.ChartType = "spline"
	pp.Title = "Sensors activities 2016"
	pp.SubTitle = "spline by sensors"
	pp.SeriesName = "Sensors"

	pp.ValueSuffix = "(°C)"
	pp.YAxisText = "Temperature (°C)"

	/* sample data is like that
	   [{ name: 'Tokyo', data: [7.0, 6.9, 9.5, 14.5, 18.2, 21.5, 25.2, 26.5, 23.3, 18.3, 13.9, 9.6] },
	    { name: 'New York', data: [-0.2, 0.8, 5.7, 11.3, 17.0, 22.0, 24.8, 24.1, 20.1, 14.1, 8.6, 2.5]},
	    .... ]
	*/

	// Prepare the data array (label/value)
	pp.DataArray = "["
	for i := range TagsList {
		pp.DataArray = pp.DataArray + "{ name: '" + TagsList[i].TagId + "', data: ["
		for e := TagsList[i].DataList.Front(); e != nil; e = e.Next() {
			t1 := e.Value.(*Estimote)
			t2 := strconv.FormatFloat(t1.MeanDistance, 'f', -1, 64)

			pp.DataArray = pp.DataArray + t2 + ","
		}
		pp.DataArray = pp.DataArray + "]},"
	}

	pp.DataArray = pp.DataArray + "]"

	fmt.Println("\n-------------\nDatas:" + pp.DataArray)

	if t, err := template.New("foo").Parse(TemplateSplineHtml); err != nil {
		log.Printf("Could not create template: %v", err)
	} else {
		if err = t.ExecuteTemplate(w, "T", pp); err != nil {
			log.Printf("Could not execute template: %v", err)
		}
	}
}

// spline,line,column,area,bar
var TemplateSplineHtml = `{{define "T"}}
<!DOCTYPE HTML>
<html>
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
        <title>GCP Sample - {{.ChartType}}</title>

        <script type="text/javascript" src="http://cdn.hcharts.cn/jquery/jquery-1.8.3.min.js"></script>
        <script type="text/javascript">
        $(function () {
            $('#container').highcharts({
                chart: {
                    // type: 'spline'
                    type: '{{.ChartType}}' // see http://api.highcharts.com/highcharts#plotOptions
                },
                title: {
                    text: '{{.Title}}',
                },
                subtitle: {
                    text: '{{.SubTitle}}',
                },
            //    xAxis: {
                  //   categories: ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12']
              //     categories: [{{.SeriesName}}] 
              //  },
                yAxis: {
                    title: {
                        // text: 'Temperature (°C)'
                        text: '{{.YAxisText}}'
                    },
                    plotLines: [{
                        value: 0,
                        width: 1,
                        color: '#808080'
                    }]
                },
                tooltip: {
                    shared: true,
                  /*  valueSuffix: '°C' */
                    valueSuffix: '{{.ValueSuffix}}'
                },
                legend: {
                    layout: 'vertical',
                    align: 'right',
                    verticalAlign: 'middle',
                    borderWidth: 0
                },
                series: {{.DataArray}}

            });
        });    
        </script>
    </head>
    <body>
    By <a id="copyright" class="anchor" href="http://www.intel.com" >olivier.guiter@intel.com</a>
    <script type="text/javascript" src="http://cdn.hcharts.cn/highcharts/4.0.1/highcharts.js"></script>
    <script type="text/javascript" src="http://cdn.hcharts.cn/highcharts/4.0.1/modules/exporting.js"></script>

    <div id="container" style="min-width: 310px; height: 400px; margin: 0 auto"></div>

    </body>
</html>
{{end}}
`

//Eof
