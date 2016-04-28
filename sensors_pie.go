// --------------------------------------------
//
// --------------------------------------------

package main

import (
	"log"
	"net/http"
	"strconv"
	"text/template"
)

/*---------------------
 */
func HandlerPIE(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/pie" {
		http.NotFound(w, r)
		return
	}

	var pp PieData
	pp.ChartType = "pie"
	pp.Title = "Sensors activities 2016"
	pp.SubTitle = "GCP sample test"
	pp.SeriesName = "Sensors"

	// Prepare the data array (label/value)
	//[[lab1, val1],[lab2, val2],...]  !!! text/html
	pp.DataArray = "["
	for i := range TagsList {
		pp.DataArray = pp.DataArray + "['" + TagsList[i].TagId + "' , " + strconv.Itoa(TagsList[i].Count) + "],"
	}
	pp.DataArray = pp.DataArray + "]"

	if t, err := template.New("foo").Parse(TemplatePieHtml); err != nil {
		log.Printf("Could not create template: %v", err)
	} else {
		if err = t.ExecuteTemplate(w, "T", pp); err != nil {
			log.Printf("Could not execute template: %v", err)
		}
	}
}

// .............
var TemplatePieHtml = `{{define "T"}}
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
                   // type: 'pie',
                    type: '{{.ChartType}}',
                    plotBackgroundColor: null,
                    plotBorderWidth: null,
                    plotShadow: false
                },
                title: {
                    text: '{{.Title}}',
                },
                subtitle: {
                    text: '{{.SubTitle}}',
                },
                tooltip: {
                    /*pointFormat: '{series.name}: <b>{point.percentage:.1f}%</b>'*/
                    pointFormat: '{series.name}: <b>{point.y}</b><br/>',
              },
                plotOptions: {
                    pie: {
                        allowPointSelect: true,
                        cursor: 'pointer',
                        dataLabels: {
                            enabled: true,
                            format: '<b>{point.name}</b>: {point.percentage:.1f} %',
                            style: {
                                color: (Highcharts.theme && Highcharts.theme.contrastTextColor) || 'black'
                            }
                        }
                    }
                },
                series: [{
                    data : {{.DataArray}}
                }]
            });
        });
		</script>
    </head>
 
    <body>
    By <a id="copyright" class="anchor" href="http://www.intel.com" >olivier.guiter@intel.com</a>
    <script type="text/javascript" src="http://cdn.hcharts.cn/highcharts/4.2.4/highcharts.js"></script>
    <script type="text/javascript" src="http://cdn.hcharts.cn/highcharts/4.2.4/modules/exporting.js"></script>

    <div id="container" style="min-width: 310px; height: 400px; max-width: 600px; margin: 0 auto"></div>

    </body>
</html>
{{end}}
`

//Eof
