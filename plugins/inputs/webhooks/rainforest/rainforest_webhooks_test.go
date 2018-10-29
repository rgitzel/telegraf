package rainforest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

func TestCurrentSummation(t *testing.T) {
    verifyGeneratedMeasurementFromMessage(t,
        CurrentSummationMessageJson(),
        map[string]interface{}{
            "delivered": "36692.711",
            "received": "4.01",
        },
        map[string]string{
            "uom": "kWh",
            "type": "CurrentSummation",
        },
        1539635859,
    )
}

func TestInstantaneousDemand(t *testing.T) {
    verifyGeneratedMeasurementFromMessage(t,
        InstantaneousDemandMessageJson(),
        map[string]interface{}{
    		"value": "0.472",
        },
        map[string]string{
            "uom": "kW",
            "type": "InstantaneousDemand",
        },
        1539632322,
    )
}

func TestMessage(t *testing.T) {
    verifyGeneratedMeasurementFromMessage(t,
        MessageMessageJson(),
        map[string]interface{}{
            "content": `{ "id": "0x00002ae0", "text": "Registration Successful", "priority": "Medium", "ConfirmationRequired": "Y", "Confirmed": "N" }`,
        },
        map[string]string{
            "type": "Message",
        },
        1539631003,
    )
}

func TestPrice(t *testing.T) {
    verifyGeneratedMeasurementFromMessage(t,
        PriceMessageJson(),
        map[string]interface{}{
    		"price": "0.0884",
        },
        map[string]string{
            "currency": "0x0348",
            "units": "USD/kWh",
            "type": "Price",
        },
        1539630914,
    )
}


func TestUnrecognized(t *testing.T) {
    verifyGeneratedMeasurementFromMessage(t,
        UnknownMessageJson(),
        map[string]interface{}{
            "message": `{ "bar": 1, "baz":"boo" }`,
        },
        map[string]string{
            "type": "unknown",
        },
        1539630914,
    )
}

func verifyGeneratedMeasurementFromMessage(t *testing.T, message string, expectedFields map[string]interface{}, expectedTags map[string]string, expectedTimestamp int64) {
	acc := postSuccessfulTestRequest(t, message)

	acc.AssertContainsMeasurement(t, DefaultMeasurementName, expectedFields, expectedTags, time.Unix(expectedTimestamp, 0))
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
