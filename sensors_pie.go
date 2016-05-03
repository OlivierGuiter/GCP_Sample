// --------------------------------------------
//
// --------------------------------------------

package main

import (
	"log"
	"net/http"
	"strconv"
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

	/*----- */

	if err := tpl.ExecuteTemplate(w, "pie.html", pp); err != nil {
		log.Printf("Could not execute template: %v", err)
	}

}
