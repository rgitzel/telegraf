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
    verifySingleMetricGenerated(t,
        CurrentSummationMessageJson(),
        mockMetric(
            map[string]string{
                "uom": "kWh",
                "type": "CurrentSummation",
            },
            map[string]interface{}{
                "delivered": 36692.711,
                "received": 4.01,
            },
            1539635859,
        ),
    )
}

func TestInstantaneousDemand(t *testing.T) {
    verifySingleMetricGenerated(t,
        InstantaneousDemandMessageJson(),
        mockMetric(
            map[string]string{
                "uom": "kW",
                "type": "InstantaneousDemand",
            },
            map[string]interface{}{
                "value": 0.472,
            },
            1539632322,
        ),
    )
}

func TestMessage(t *testing.T) {
    verifySingleMetricGenerated(t,
        MessageMessageJson(),
        mockMetric(
            map[string]string{
                "type": "Message",
            },
            map[string]interface{}{
                "content": `{ "id": "0x00002ae0", "text": "Registration Successful", "priority": "Medium", "ConfirmationRequired": "Y", "Confirmed": "N" }`,
            },
            1539631003,
        ),
    )
}

func TestPrice(t *testing.T) {
    verifySingleMetricGenerated(t,
        PriceMessageJson(),
        mockMetric(
            map[string]string{
                "currency": "0x0348",
                "units": "USD/kWh",
                "type": "Price",
            },
            map[string]interface{}{
                "price": 0.0884,
            },
            1539630914,
        ),
    )
}


func TestValidButUnrecognized(t *testing.T) {
    verifySingleMetricGenerated(t,
        ValidButUnrecognizedMessageJson(),
        mockMetric(
            map[string]string{
                "type": "unknown",
            },
            map[string]interface{}{
                "message": `{ "bar": 1, "baz": "boo" }`,
            },
            1539630914,
        ),
    )
}


func TestNone(t *testing.T) {
    verifyNoMetricsGenerated(t, EmptyMessagesListJson())
}

func TestMultiple(t *testing.T) {
    verifyMultipleMetricsGenerated(t,
        MultipleMessagesJson(),
        []*testutil.Metric{
            mockMetric(
                map[string]string{
                    "uom": "kW",
                    "type": "InstantaneousDemand",
                },
                map[string]interface{}{
                    "value": 0.2,
                },
                1539632324,
            ),
            mockMetric(
                map[string]string{
                    "uom": "kW",
                    "type": "InstantaneousDemand",
                },
                map[string]interface{}{
                    "value": 0.472,
                },
                1539632322,
            ),
            mockMetric(
                map[string]string{
                    "uom": "kWh",
                    "type": "CurrentSummation",
                },
                map[string]interface{}{
                    "delivered": 36692.711,
                    "received": 4.01,
                },
                153963589,
            ),
        },
    )
}


func TestNotJson(t *testing.T) {
    verifyNoMetricsGenerated(t, "lkasjdlf sdlfaksdf")
}

func TestJsonNotEnvelope(t *testing.T) {
    verifyNoMetricsGenerated(t, `{"foo": 1}`)
}

func mockMetric(expectedTags map[string]string, expectedFields map[string]interface{}, expectedMillis int64) *testutil.Metric {
    return &testutil.Metric{DefaultMeasurementName, expectedTags, expectedFields, time.Unix(expectedMillis, 0)}
}

func verifyNoMetricsGenerated(t *testing.T, message string) {
    verifyMultipleMetricsGenerated(t, message, []*testutil.Metric{})
}

func verifySingleMetricGenerated(t *testing.T, message string, metric *testutil.Metric) {
    verifyMultipleMetricsGenerated(t, message, []*testutil.Metric{metric})
}

func verifyMultipleMetricsGenerated(t *testing.T, message string, metrics []*testutil.Metric) {
	acc := postSuccessfulTestRequest(t, message)
	acc.AssertContainsMetrics(t, metrics)
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
