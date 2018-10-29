package rainforest


func InstantaneousDemandMessageJson() string {
    return `{
              "timestamp":"1539632334000",
              "deviceGuid":"d8d5b9aa2a",
              "body": [{
              "timestamp":"1539632322000",
              "subdeviceGuid":"000780e67e73",
              "componentId":"all",
              "dataType":"InstantaneousDemand",
              "data":{
                "demand":0.472,
                "units":"kW"
              }}
              ]
            }`
}

func CurrentSummationMessageJson() string {
    return `{
              "timestamp":"1539635873000",
              "deviceGuid":"d8d5b9aa2a",
              "body": [{
              "timestamp":"1539635859000",
              "subdeviceGuid":"000780e67e73",
              "componentId":"all",
              "dataType":"CurrentSummation",
              "data":{
                "summationDelivered":36692.711,
                "summationReceived":4.01,
                "units":"kWh"
              }}
              ]
            }`
}

func RegistrationSuccessfulMessageJson() string {
    return `{
              "timestamp":"1539631913000",
              "deviceGuid":"d8d5b9aa2a",
              "body": [{
              "timestamp":"1539631884000",
              "subdeviceGuid":"000780e67e73",
              "componentId":"all",
              "dataType":"Message",
              "data":{
                "id": "0x00002ae0",
                "text": "Registration Successful",
                "priority": "Medium"
              }}
              ]
            }`
}

function PriceMessageJson() string {
    return `{
              "timestamp":"1539630998000",
              "deviceGuid":"d8d5b9aa2a",
              "body": [{
              "timestamp":"1539630914000",
              "subdeviceGuid":"00078e67e73",
              "componentId":"all",
              "dataType":"Price",
              "data":{
                "price":0.0884,
                "currency": "0x0348",
                "units":"USD/kWh"
              }}
              ]
            }`
}


function MessageMessageJson() string {
    return `{
              "timestamp":"1539631037000",
              "deviceGuid":"d8d5b9aa2a",
              "body": [{
              "timestamp":"1539631003000",
              "subdeviceGuid":"00078e67e73",
              "componentId":"all",
              "dataType":"Message",
              "data":{
                "id": "0x00002ae0",
                "text": "Registration Successful",
                "priority": "Medium",
                "ConfirmationRequired": "Y",
                "Confirmed": "N"
              }}
              ]
            }`
}