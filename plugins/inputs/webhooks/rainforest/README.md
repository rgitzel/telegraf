# Rainforest webhooks

This plugin allows Telegraf to receive messages uploaded by a
[Rainforest](https://rainforestautomation.com)
energy monitoring gateway.

The following Rainforest devices are supported:
- [Eagle 200](https://rainforestautomation.com/rfa-z114-eagle-200-2/)
  - version 2.2 Uploader API (the specification PDF can be found [here](https://rainforestautomation.com/support/developer/))
  - "RFA" (i.e. JSON) messages only
  - Streaming Mode
  - Buffered Mode probably works, not yet tested
    - but how configure this? not via the webapp

# Architecture

TBD. Include a diagram.


# How to Use the Plugin

## Add an Upload Destination

## Record Raw Data

The plugin attempts to record everything received from the Eagle, and writes it into a
single InfluxDb measurement.  You can specify a name for this measurement, otherwise it
will default to `rainforest_eagle_200_raw`

**TODO:** where configure retention?

## Optionally Extract Useful Data

While you can certainly use only the raw data measurement, it may take up more
space than you would like.  One option is to put a short retention period on the
raw data measurement (one or two weeks, say) and use a continuous query or Kapacitor
**thing** to extract the actually useful bits into a simpler form.

TBD:  examples


## Graph the Data

TBD:  Grafana examples


# Supported Messages

## Instantaneous Demand

This provides the amount of power being drawn a specific point of time, in kiloWatts.

**Tags:**
- `type` - string = "InstantaneousDemand"
- `uom` - string = "kW"

**Fields:**
- `value` - float

**TODO:** distinguish between delivered and received?


## Current Summation

This provides the total power "delivered" to and "received" from you (in kilowatt-hours) since your meter was connected. Calculated the difference
that value at two different times will provide the amounta of power consumed and returned over that time period.

The "received" value will always be zero in most cases, unless say you have solar panels
feed power back into your utility.

Note that successive messages typically do not have a changed value; the value seems to change only every
minute or two, and messages are sent more frequently than that.  **TODO:** confirm with vendor.

**Tags:**
- `type` - string = "CurrentSummation"
- `uom` - string = "kWh"

**Fields:**
- `delivered` - float - amount of power from utility to user
- `received` - float- amount of power from user to utility


## Price

This provides the utility's current price for power.

If your utility does not change the
price for electricity throughout the day, then you can ignore this, the value will very
rarely change.

**Tags:**
- `currency` - string - "0x0348"
- `uom` = "USD/kWh"

**Fields:**
- `price` - float

**TODO:** why both 'currency' and 'units'?


## Message

**Tags:**
- `type` = "Message"

**Fields:**
- `content` - string - the JSON contents of the message


## Others

Any messages received other than the above will simply be recorded.

**Tags:**
- `type` = "Unrecognized"

**Fields:**
- `content` - string - the JSON contents of the message





