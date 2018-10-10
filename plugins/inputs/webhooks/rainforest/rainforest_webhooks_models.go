package rainforest

import (
	"encoding/json"
)

/*
 * an envelope contains a list of messages; each message contains a block of data
 *
 * each message has its own timestamp, and then the envelope has one as well,
 *  which should always be later than the message timestamps
 */

type RainforestEnvelope struct {
	Timestamp   int64               `json:"timestamp"`
    Messages    []RainforestMessage `json:"body"`
}

type RainforestMessage struct {
	Timestamp   string          `json:"timestamp"`
    Type        string          `json:"dataType"`
    RawData     json.RawMessage `json:"data"`
}

type RainforestCurrentSummationData struct {
    Delivered   json.Number  `json:"summationDelivered"`
    Received    json.Number  `json:"summationReceived"`
    Units       string  `json:"units"`
}

type RainforestInstantaneousDemandData struct {
    Value       json.Number  `json:"demand"`
    Units       string  `json:"units"`
}
