package rainforest

import (
	"encoding/json"
)

/*
 * an 'envelope' contains a list of 'messages'; each message contains a block
 *  of data specific to the "type" of the particular message
 *
 * each message has its own timestamp, and the envelope has one as well;
 *  the envelope timestamp should always be later than all of the messages' timestamps
 */

type RainforestEnvelope struct {
	Timestamp   int64               `json:"timestamp"`
	Messages    []RainforestMessage `json:"body"`
}

type RainforestMessage struct {
	Timestamp   string              `json:"timestamp"`
	Type        string              `json:"dataType"`
	RawData	    json.RawMessage     `json:"data"`
}

/*
 * the rest of these each match the 'body.data' node
 *  for a particular type of message
 */

type RainforestCurrentSummationData struct {
	Delivered   json.Number         `json:"summationDelivered"`
	Received    json.Number         `json:"summationReceived"`
	Units	    string              `json:"units"`
}

type RainforestInstantaneousDemandData struct {
	Value	    json.Number         `json:"demand"`
	Units	    string              `json:"units"`
}

type RainforestPriceData struct {
	Price	    json.Number         `json:"price"`
	Currency    string              `json:"currency"`
	Units	    string              `json:"units"`
}
