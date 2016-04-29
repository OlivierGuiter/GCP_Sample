// --------------------------------------------
// Code sample:
// Get data from GCP PubSub
// Display charts
// --------------------------------------------

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/cloud/pubsub"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"io/ioutil"

	"container/list"
)

// Estimote data struct
type GpsReadout struct {
	Latitude  float64 `json:"lat,string"`
	LatPole   string  `json:"latPole"`
	Longitude float64 `json:"lon,string"`
	LonPole   string  `json:"lonPole"`
}

type TlmData struct {
	Version uint8   `json:"version"`
	Vbatt   uint16  `json:"vbatt"`
	Temp    float32 `json:"temp"`
	AdvCnt  uint32  `json:"advCnt"`
	SecCnt  uint32  `json:"secCnt"`
}

type Estimote struct {
	TagId             string     `json:"tagId"`
	GatewayId         string     `json:"gatewayId"`
	GatewayLocation   string     `json:"gatewayLocation"`
	GatewayGpsReadout GpsReadout `json:"gatewayGpsReadout"`
	LastSeen          int64      `json:"lastSeen"`
	SmoothingWindow   int        `json:"smoothingWindow"`
	Rssi	          int        `json:"rssi"`
	Distance	  float64    `json:"distance"`
	LowerDistance     float64    `json:"lowerDistance"`
	MeanDistance      float64    `json:"meanDistance"`
	UpperDistance     float64    `json:"upperDistance"`

	Tlm TlmData "tlm"
}

type TagEntry struct {
	TagId    string
	Count    int
	DataList *list.List // Store the last MaxDataKeep (100)
}

type PieData struct {
	Title      string
	ChartType  string
	SubTitle   string
	SeriesName string
	DataArray  string

	//SPline
	YAxisText   string
	ValueSuffix string
}

var (
	count   int
	countMu sync.Mutex
	ctx     context.Context

	TagsListMu sync.Mutex
	TagsList   []TagEntry
)

const (
	ProjID      = "sensor-it"   // could be get from os.Getenv[GCLOUD_PROJECT")
	SubTopic    = "sensorTopic" // could ... PUBSUB_TOPIC
	SubName     = "sensorSub"   // could ... PUBSUB_SUB
	MaxDataKeep = 100

	KeyFileJSON = "sensor-it-7021d8976ff0.json"
)

func main() {

	// Get the context
	ctx := cloudContext(ProjID)

	// a new client ...
	client, err := pubsub.NewClient(ctx, ProjID)
	if err != nil {
		log.Fatalf("creating pubsub client: %v", err)
	}

	// Create topic (check if exist...)
	exists, err := client.Topic(SubTopic).Exists(ctx)
	if err != nil {
		log.Fatalf("Checking topic exists failed: %v", err)
	}
	if exists {
		fmt.Printf("Topic %s is already created.\n", SubTopic)
	} else {
		_, err = client.NewTopic(ctx, SubTopic)
		if err != nil {
			log.Fatalf("Creating topic failed: %v", err)
		}
		fmt.Printf("Topic %s was created.\n", SubTopic)
	}

	// Create subscription
	exists, err = client.Subscription(SubName).Exists(ctx)
	if err != nil {
		log.Fatalf("Checking subscription exists failed: %v", err)
	}

	if exists {
		fmt.Printf("Subscription %s is already there.\n", SubName)
	} else {
		_, err = client.NewSubscription(ctx, SubName, client.Topic(SubTopic), time.Duration(0), nil)

		if err != nil {
			log.Fatalf("Creating Subscription failed: %v", err)
		}
		fmt.Printf("Subscription %s was created.\n", SubName)
	}

	/* handler dispatcher */
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	http.HandleFunc("/_ah/stop", shutdownHandler)
	http.HandleFunc("/pie", HandlerPIE)
	http.HandleFunc("/distance", HandlerDistance)
	http.HandleFunc("/test", HandlerTEST)

	go subscribe(ctx)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

/* ---------------------------------------------
 * Thread sub: Wait for incoming messages from sensors
 * thru the pubsub channel
 */
func subscribe(ctx context.Context) {
	log.Printf("In subscribe (ctx: %+v)!!!\n", ctx)
	for {
		// Pull up to 10 messages (maybe fewer) from the subscription.
		// Blocks for an indeterminate amount of time.
		msgs, err := pubsub.PullWait(ctx, SubName, 10)
		if err != nil {
			log.Fatalf("could not pull: %v", err)
		}

		for _, m := range msgs {
			msg := m

			var estimote Estimote

			if err := json.Unmarshal(msg.Data, &estimote); err != nil {
				log.Printf("could not decode message data: %s \n%v", msg.Data, err)
				go pubsub.Ack(ctx, SubName, msg.AckID)
				continue
			} else {
fmt.Println("\n-------------")
				log.Printf("Debug data: %s\n", msg.Data)
				log.Printf("Debug gw loc: %#v\n\n", estimote)

			}

			//		log.Printf("[ID %s] Processing.", estimote.TagId)
			go func() {
				if err := updateData(&estimote); err != nil {
					log.Printf("[ID %d] could not update: %v", estimote.TagId, err)
					return
				}

				countMu.Lock()
				count++
				countMu.Unlock()

				pubsub.Ack(ctx, SubName, msg.AckID)
				//log.Printf("[ID %d] ACK", estimote.TagId)
			}()
		}
	}

}

/* Update the Sensors array
 * count, for each sensor, the number of incoming frames
 * and, in another array, store the last 10 records
 */
func updateData(estimote *Estimote) error {
	//log.Printf("UPDATE [ID %s]\n", estimote.TagId)
	var bfound bool

	bfound = false
	for idx := range TagsList {
		if TagsList[idx].TagId == estimote.TagId {
			//log.Printf("Found [ID %s]: %d \n", estimote.TagId, TagsList[idx].Count)
			TagsListMu.Lock()
			TagsList[idx].Count++
			TagsList[idx].DataList.PushBack(estimote)

			if TagsList[idx].DataList.Len() > MaxDataKeep {
				TagsList[idx].DataList.Remove(TagsList[idx].DataList.Front())
				//	log.Printf("Queue Len [ID %s]: %d\n", estimote.TagId, TagsList[idx].DataList.Len())
			}
			TagsListMu.Unlock()
			bfound = true
		}
	}
	// if not found, add the entry
	if bfound == false {
		var esttmp TagEntry
		esttmp.TagId = estimote.TagId
		esttmp.Count = 0
		//Create the data ring buffer
		esttmp.DataList = list.New()
		esttmp.DataList.PushFront(estimote)
		TagsListMu.Lock()
		TagsList = append(TagsList, esttmp)
		TagsListMu.Unlock()
		log.Printf("Append to list [ID %s]\n", estimote.TagId)
		fmt.Println("%#v", estimote)
	}

	return nil
}

/*
 */
func HandlerTEST(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Shutdown app\n")
	for i := range TagsList {
		fmt.Println("-----------------------")
		e := TagsList[i].DataList.Front()
		// do something with e.Value
		fmt.Println("%#v", e.Value)
		fmt.Println("%+v", e.Value)
		fmt.Println("%T", e.Value)
	}
}

/*
 */
func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Shutdown app\n")
}

/*
 */
func rootHandler(w http.ResponseWriter, r *http.Request) {
	TagsListMu.Lock()
	defer TagsListMu.Unlock()

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := tmpl.Execute(w, TagsList); err != nil {
		log.Printf("Could not execute template: %v", err)
	}
}

/*
 */
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

/*
 */
func cloudContext(projectID string) context.Context {

	// Get the key file
	jsonKey, err := ioutil.ReadFile(KeyFileJSON)
	if err != nil {
		log.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		pubsub.ScopeCloudPlatform,
		pubsub.ScopePubSub,
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := cloud.NewContext(projectID, conf.Client(oauth2.NoContext))

	return ctx

}

// Very simple template to display sensors various data
var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
  <head>
    <title>Sensor IT</title>
  </head>
  <body>
    <div>
      <p>Tag ID list received by this instance:</p>
      <ul>
    	{{range .}}
<li>Tag ID: {{.TagId}} / msgs received: {{.Count}}</li>
    	{{end}}
      </ul>
    </div>
    <p>Note: if the application is running across multiple instances, each
      instance will have its own list of tag ID.</p>
  </body>
</html>`))

//Eof
