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
    verifyGeneratedMetricFromMessage(t,
        CurrentSummationMessageJson(),
        map[string]string{
            "uom": "kWh",
            "type": "CurrentSummation",
        },
        map[string]interface{}{
            "delivered": 36692.711,
            "received": 4.01,
        },
        1539635859,
    )
}

func TestInstantaneousDemand(t *testing.T) {
    verifyGeneratedMetricFromMessage(t,
        InstantaneousDemandMessageJson(),
        map[string]string{
            "uom": "kW",
            "type": "InstantaneousDemand",
        },
        map[string]interface{}{
    		"value": 0.472,
        },
        1539632322,
    )
}

func TestMessage(t *testing.T) {
    verifyGeneratedMetricFromMessage(t,
        MessageMessageJson(),
        map[string]string{
            "type": "Message",
        },
        map[string]interface{}{
            "content": `{ "id": "0x00002ae0", "text": "Registration Successful", "priority": "Medium", "ConfirmationRequired": "Y", "Confirmed": "N" }`,
        },
        1539631003,
    )
}

func TestPrice(t *testing.T) {
    verifyGeneratedMetricFromMessage(t,
        PriceMessageJson(),
        map[string]string{
            "currency": "0x0348",
            "units": "USD/kWh",
            "type": "Price",
        },
        map[string]interface{}{
    		"price": 0.0884,
        },
        1539630914,
    )
}


func TestUnrecognized(t *testing.T) {
    verifyGeneratedMetricFromMessage(t,
        UnknownMessageJson(),
        map[string]string{
            "type": "unknown",
        },
        map[string]interface{}{
            "message": `{ "bar": 1, "baz":"boo" }`,
        },
        1539630914,
    )
}

func verifyGeneratedMetricFromMessage(t *testing.T, message string, expectedTags map[string]string, expectedFields map[string]interface{}, expectedTimestamp int64) {
	acc := postSuccessfulTestRequest(t, message)

    acc.AssertMetricsCount(t, 1)
	acc.AssertContainsMetric(t, DefaultMeasurementName, expectedFields, expectedTags, time.Unix(expectedTimestamp, 0))
}

func postSuccessfulTestRequest(t *testing.T, json string) testutil.Accumulator {
	w := httptest.NewRecorder()

	var acc testutil.Accumulator
	wh := &RainforestWebhook{acc: &acc}

	req, _ := http.NewRequest("POST", "/rainforest", strings.NewReader(json))
	wh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}

	return acc
}
