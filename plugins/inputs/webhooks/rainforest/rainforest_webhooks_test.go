package rainforest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

func TestInstantaneousDemand(t *testing.T) {
    expectedFields := map[string]interface{}{
        "value": "0.472",
    }

    expectedTags := map[string]string{
        "uom": "kW",
        "type": "InstantaneousDemand",
    }

    acc := postSuccessfulTestRequest(t, InstantaneousDemandMessageJson())

	acc.AssertContainsMeasurement(t, MeasurementName, expectedFields, expectedTags, time.Unix(1539632322, 0))
}

func TestCurrentSummation(t *testing.T) {
    expectedFields := map[string]interface{}{
        "delivered": "36692.711",
        "received": "4.01",
    }

    expectedTags := map[string]string{
        "uom": "kWh",
        "type": "CurrentSummation",
    }

    acc := postSuccessfulTestRequest(t, CurrentSummationMessageJson())

	acc.AssertContainsMeasurement(t, MeasurementName, expectedFields, expectedTags, time.Unix(1539635859, 0))
}


func TestUnrecognized(t *testing.T) {
    expectedFields := map[string]interface{}{
        "message": "oops",//[]byte(`{'summationDelivered':36692.711,'summationReceived':4.01,'units':'kWh'}`),
    }

    expectedTags := map[string]string{
        "type": "unknown",
    }

    acc := postSuccessfulTestRequest(t, RegistrationSuccessfulMessageJson())

	acc.AssertContainsMeasurement(t, MeasurementName, expectedFields, expectedTags, time.Unix(1539631884, 0))
}

func postSuccessfulTestRequest(t *testing.T, json string) testutil.Accumulator {

	req, _ := http.NewRequest("POST", "/rainforest", strings.NewReader(json))

	var acc testutil.Accumulator
	wh := &RainforestWebhook{acc: &acc}

	w := httptest.NewRecorder()
	wh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}

	return acc
}