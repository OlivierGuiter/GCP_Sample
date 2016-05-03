// --------------------------------------------
//
// --------------------------------------------

package main

import (
	//	"fmt"
	"log"
	"net/http"
	"strconv"
)

/*---------------------
 */
func HandlerDistance(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/distance" {
		http.NotFound(w, r)
		return
	}

	var pp PieData
	pp.ChartType = "spline"
	pp.Title = "Sensors activities 2016"
	pp.SubTitle = "Sensors range"
	pp.SeriesName = "Sensors"

	pp.ValueSuffix = "dist"
	pp.YAxisText = "Approximate distance from beacon"

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
			//			t2 := strconv.FormatFloat(float64(t1.Tlm.Temp), 'f', -1, 32)
			t2 := strconv.FormatFloat(t1.MeanDistance, 'f', -1, 64)
			//			t2 := strconv.Itoa(t1.Rssi * -1)
			pp.DataArray = pp.DataArray + t2 + ","
		}
		pp.DataArray = pp.DataArray + "]},"
	}

	pp.DataArray = pp.DataArray + "]"

	//	fmt.Println("\n-------------\nDatas:" + pp.DataArray)

	if err := tpl.ExecuteTemplate(w, "spline.html", pp); err != nil {
		log.Printf("Could not execute template: %v", err)
	}

}

//Eof
