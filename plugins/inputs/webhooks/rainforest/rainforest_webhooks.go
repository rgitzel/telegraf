package rainforest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

const MeasurementName = "rainforest_eagle_200_raw"


type RainforestWebhook struct {
	Path   string
	acc	telegraf.Accumulator
}

func (rw *RainforestWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(rw.Path, rw.eventHandler).Methods("POST")
	log.Printf("I! Started the webhooks_rainforest on %s\n", rw.Path)
	rw.acc = acc
}

func (rw *RainforestWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var envelope RainforestEnvelope
	json.Unmarshal(contents, &envelope)

	for _, message := range envelope.Messages {
		rw.possibleAddMeasurement(MeasurementName, &message)
	}

	w.WriteHeader(http.StatusOK)
}

func (rw *RainforestWebhook) possibleAddMeasurement(name string, message *RainforestMessage) {
	var tags = map[string]string {
		"type": message.Type,
	}
	var fields = make(map[string]interface{})

	parsed, _ := strconv.ParseInt(message.Timestamp, 10, 64)
	var ts = time.Unix(parsed / 1000, 0)

	if message.Type == "InstantaneousDemand" {
		var data RainforestInstantaneousDemandData
		json.Unmarshal(message.RawData, &data)

		tags["uom"] = data.Units

		fields["value"] = string(data.Value)

	} else if message.Type == "CurrentSummation" {
		var data RainforestCurrentSummationData
		json.Unmarshal(message.RawData, &data)

		tags["uom"] = data.Units

		fields["delivered"] = string(data.Delivered)
		fields["received"] = string(data.Received)

	} else {
		tags["type"] = "unknown"

		//fields["message"], _ = json.Marshal(message.RawData)
		fields["message"] = "oops"
	}

	log.Printf("processed message type '%s'", message.Type)

	rw.acc.AddFields(name, fields, tags, ts)
}
