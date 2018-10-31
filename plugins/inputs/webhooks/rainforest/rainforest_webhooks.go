package rainforest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

const DefaultMeasurementName = "rainforest_eagle_200_raw"


type RainforestWebhook struct {
	Path string
	Name string
	acc	 telegraf.Accumulator
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
		rw.possiblyAddMeasurement(&message)
	}

	w.WriteHeader(http.StatusOK)
}

func (rw *RainforestWebhook) possiblyAddMeasurement(message *RainforestMessage) {
	var tags = map[string]string {
		"type": message.Type,
	}
	var fields = make(map[string]interface{})

	parsed, _ := strconv.ParseInt(message.Timestamp, 10, 64)
	var ts = time.Unix(parsed / 1000, 0)

	switch t := message.Type; t {
		case "CurrentSummation":
			addCurrentSummationValues(message.RawData, tags, fields)

		case "InstantaneousDemand":
			addInstantaneousDemandValues(message.RawData, tags, fields)

		case "Message":
			addMessageValues(message.RawData, tags, fields)

		case "Price":
			addPriceValues(message.RawData, tags, fields)

		default:
			addUnknownValues(message.RawData, tags, fields)
	}

	log.Printf("processed message type '%s'", message.Type)

	// I would rather set this in `Register()` but then it ends up over-written?
	if (rw.Name == "") {
		rw.Name = DefaultMeasurementName
	}

	rw.acc.AddFields(rw.Name, fields, tags, ts)
}

func addCurrentSummationValues(dataNode json.RawMessage, tags map[string]string, fields map[string]interface{}) {
	var data RainforestCurrentSummationData
	json.Unmarshal(dataNode, &data)

	tags["uom"] = data.Units

	fields["delivered"], _ = data.Delivered.Float64()
	fields["received"], _ = data.Received.Float64()
}

func addInstantaneousDemandValues(dataNode json.RawMessage, tags map[string]string, fields map[string]interface{}) {
	var data RainforestInstantaneousDemandData
	json.Unmarshal(dataNode, &data)

	tags["uom"] = data.Units

	fields["value"], _ = data.Value.Float64()
}

func addMessageValues(dataNode json.RawMessage, tags map[string]string, fields map[string]interface{}) {
	fields["content"] = cleanRawJson(dataNode)
}

func addPriceValues(dataNode json.RawMessage, tags map[string]string, fields map[string]interface{}) {
	var data RainforestPriceData
	json.Unmarshal(dataNode, &data)

	tags["currency"] = data.Currency
	tags["units"] = data.Units

	fields["price"], _ = data.Price.Float64()
}

func addUnknownValues(dataNode json.RawMessage, tags map[string]string, fields map[string]interface{}) {
	tags["type"] = "unknown"

	fields["message"] = cleanRawJson(dataNode)
}

// the closest I can seem to find to getting a nice clean one-line version of a pretty JSON string :(
func cleanRawJson(json json.RawMessage) string {
    r := regexp.MustCompile("\n|\t| {2,}")
	return r.ReplaceAllString(r.ReplaceAllString(string(json), " "), " ")
}