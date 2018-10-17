package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/assert"
)

// Metric defines a single point measurement
type Metric struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Time        time.Time
}

func (p *Metric) String() string {
	return fmt.Sprintf("%s %v %v %d", p.Measurement, p.Fields, p.Tags, p.Time.Unix())
}

// Accumulator defines a mocked out accumulator
type Accumulator struct {
	sync.Mutex
	*sync.Cond

	Metrics  []*Metric
	nMetrics uint64
	Discard  bool
	Errors   []error
	debug	bool
}

func (a *Accumulator) Strings() []string {
	list := []string{}
	for _, m := range a.Metrics {
		list = append(list, m.String())
	}
	return list
}

func (a *Accumulator) NMetrics() uint64 {
	return atomic.LoadUint64(&a.nMetrics)
}

func (a *Accumulator) FirstError() error {
	if len(a.Errors) == 0 {
		return nil
	}
	return a.Errors[0]
}

func (a *Accumulator) ClearMetrics() {
	a.Lock()
	defer a.Unlock()
	atomic.StoreUint64(&a.nMetrics, 0)
	a.Metrics = make([]*Metric, 0)
}

func (a *Accumulator) Contains(m *Metric) bool {
	for _, candidate := range a.Metrics {
		if reflect.DeepEqual(candidate, m) {
			return true
		}
	}
	return false
}

// AddFields adds a measurement point with a specified timestamp.
func (a *Accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp ...time.Time,
) {
	a.Lock()
	defer a.Unlock()
	atomic.AddUint64(&a.nMetrics, 1)
	if a.Cond != nil {
		a.Cond.Broadcast()
	}
	if a.Discard {
		return
	}

	tagsCopy := map[string]string{}
	for k, v := range tags {
		tagsCopy[k] = v
	}

	fieldsCopy := map[string]interface{}{}
	for k, v := range fields {
		fieldsCopy[k] = v
	}

	if len(fields) == 0 {
		return
	}

	var t time.Time
	if len(timestamp) > 0 {
		t = timestamp[0]
	} else {
		t = time.Now()
	}

	if a.debug {
		pretty, _ := json.MarshalIndent(fields, "", "  ")
		prettyTags, _ := json.MarshalIndent(tags, "", "  ")
		msg := fmt.Sprintf("Adding Measurement [%s]\nFields:%s\nTags:%s\n",
			measurement, string(pretty), string(prettyTags))
		fmt.Print(msg)
	}

	p := &Metric{
		Measurement: measurement,
		Fields:      fields,
		Tags:        tagsCopy,
		Time:        t,
	}

	a.Metrics = append(a.Metrics, p)
}

func (a *Accumulator) AddCounter(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp ...time.Time,
) {
	a.AddFields(measurement, fields, tags, timestamp...)
}

func (a *Accumulator) AddGauge(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp ...time.Time,
) {
	a.AddFields(measurement, fields, tags, timestamp...)
}

func (a *Accumulator) AddMetrics(metrics []telegraf.Metric) {
	for _, m := range metrics {
		a.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
}

func (a *Accumulator) AddSummary(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp ...time.Time,
) {
	a.AddFields(measurement, fields, tags, timestamp...)
}

func (a *Accumulator) AddHistogram(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp ...time.Time,
) {
	a.AddFields(measurement, fields, tags, timestamp...)
}

// AddError appends the given error to Accumulator.Errors.
func (a *Accumulator) AddError(err error) {
	if err == nil {
		return
	}
	a.Lock()
	a.Errors = append(a.Errors, err)
	if a.Cond != nil {
		a.Cond.Broadcast()
	}
	a.Unlock()
}

func (a *Accumulator) SetPrecision(precision, interval time.Duration) {
	return
}

func (a *Accumulator) DisablePrecision() {
	return
}

func (a *Accumulator) Debug() bool {
	// stub for implementing Accumulator interface.
	return a.debug
}

func (a *Accumulator) SetDebug(debug bool) {
	// stub for implementing Accumulator interface.
	a.debug = debug
}

// Get gets the specified measurement point from the accumulator
func (a *Accumulator) Get(measurement string) (*Metric, bool) {
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			return p, true
		}
	}

	return nil, false
}

func (a *Accumulator) HasTag(measurement string, key string) bool {
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			_, ok := p.Tags[key]
			return ok
		}
	}
	return false
}

func (a *Accumulator) TagValue(measurement string, key string) string {
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			v, ok := p.Tags[key]
			if !ok {
				return ""
			}
			return v
		}
	}
	return ""
}

// Calls the given Gather function and returns the first error found.
func (a *Accumulator) GatherError(gf func(telegraf.Accumulator) error) error {
	if err := gf(a); err != nil {
		return err
	}
	if len(a.Errors) > 0 {
		return a.Errors[0]
	}
	return nil
}

// NFields returns the total number of fields in the accumulator, across all
// measurements
func (a *Accumulator) NFields() int {
	a.Lock()
	defer a.Unlock()
	counter := 0
	for _, pt := range a.Metrics {
		for range pt.Fields {
			counter++
		}
	}
	return counter
}

// Wait waits for the given number of metrics to be added to the accumulator.
func (a *Accumulator) Wait(n int) {
	a.Lock()
	if a.Cond == nil {
		a.Cond = sync.NewCond(&a.Mutex)
	}
	for int(a.NMetrics()) < n {
		a.Cond.Wait()
	}
	a.Unlock()
}

// WaitError waits for the given number of errors to be added to the accumulator.
func (a *Accumulator) WaitError(n int) {
	a.Lock()
	if a.Cond == nil {
		a.Cond = sync.NewCond(&a.Mutex)
	}
	for len(a.Errors) < n {
		a.Cond.Wait()
	}
	a.Unlock()
}

func (a *Accumulator) AssertContainsTaggedFields(
	t *testing.T,
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if !reflect.DeepEqual(tags, p.Tags) {
			continue
		}

		if p.Measurement == measurement {
			assert.Equal(t, fields, p.Fields)
			return
		}
	}
	msg := fmt.Sprintf("unknown measurement %s with tags %v", measurement, tags)
	assert.Fail(t, msg)
}

func (a *Accumulator) AssertDoesNotContainsTaggedFields(
	t *testing.T,
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if !reflect.DeepEqual(tags, p.Tags) {
			continue
		}

		if p.Measurement == measurement {
			assert.Equal(t, fields, p.Fields)
			msg := fmt.Sprintf("found measurement %s with tags %v which should not be there", measurement, tags)
			assert.Fail(t, msg)
		}
	}
	return
}

func (a *Accumulator) AssertMeasurementsCount(t *testing.T, expectedCount int) {
	a.Lock()
	defer a.Unlock()

	assert.Equal(t, expectedCount, len(a.Metrics))
}

func (a *Accumulator) AssertContainsMeasurement(
	t *testing.T,
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp time.Time,
) {
	a.Lock()
	defer a.Unlock()

	if len(a.Metrics) == 0 {
		assert.Fail(t, "accumulator is empty")
	} else {
		n := Metric{measurement, tags, fields, timestamp}
		if a.Contains(&n) {
			return
		}
		accumulatorString := strings.Join(a.Strings()[:], "\n")
		msg := fmt.Sprintf("measurement '%s' not found in the accumulator:\n%s", n.String(), accumulatorString)
		assert.Fail(t, msg)
	}
}

func (a *Accumulator) AssertContainsFields(
	t *testing.T,
	measurement string,
	fields map[string]interface{},
) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			assert.Equal(t, fields, p.Fields)
			return
		}
	}
	msg := fmt.Sprintf("unknown measurement %s", measurement)
	assert.Fail(t, msg)
}

func (a *Accumulator) HasPoint(
	measurement string,
	tags map[string]string,
	fieldKey string,
	fieldValue interface{},
) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement != measurement {
			continue
		}

		if !reflect.DeepEqual(tags, p.Tags) {
			continue
		}

		v, ok := p.Fields[fieldKey]
		if ok && reflect.DeepEqual(v, fieldValue) {
			return true
		}
	}
	return false
}

func (a *Accumulator) AssertDoesNotContainMeasurement(t *testing.T, measurement string) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			msg := fmt.Sprintf("found unexpected measurement %s", measurement)
			assert.Fail(t, msg)
		}
	}
}

// HasTimestamp returns true if the measurement has a matching Time value
func (a *Accumulator) HasTimestamp(measurement string, timestamp time.Time) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			return timestamp.Equal(p.Time)
		}
	}

	return false
}

// HasField returns true if the given measurement has a field with the given
// name
func (a *Accumulator) HasField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			if _, ok := p.Fields[field]; ok {
				return true
			}
		}
	}

	return false
}

// HasIntField returns true if the measurement has an Int value
func (a *Accumulator) HasIntField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(int)
					return ok
				}
			}
		}
	}

	return false
}

// HasInt64Field returns true if the measurement has an Int64 value
func (a *Accumulator) HasInt64Field(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(int64)
					return ok
				}
			}
		}
	}

	return false
}

// HasInt32Field returns true if the measurement has an Int value
func (a *Accumulator) HasInt32Field(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(int32)
					return ok
				}
			}
		}
	}

	return false
}

// HasStringField returns true if the measurement has an String value
func (a *Accumulator) HasStringField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(string)
					return ok
				}
			}
		}
	}

	return false
}

// HasUIntField returns true if the measurement has a UInt value
func (a *Accumulator) HasUIntField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(uint64)
					return ok
				}
			}
		}
	}

	return false
}

// HasFloatField returns true if the given measurement has a float value
func (a *Accumulator) HasFloatField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(float64)
					return ok
				}
			}
		}
	}

	return false
}

// HasMeasurement returns true if the accumulator has a measurement with the
// given name
func (a *Accumulator) HasMeasurement(measurement string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			return true
		}
	}
	return false
}

// IntField returns the int value of the given measurement and field or false.
func (a *Accumulator) IntField(measurement string, field string) (int, bool) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					v, ok := value.(int)
					return v, ok
				}
			}
		}
	}

	return 0, false
}

// Int64Field returns the int64 value of the given measurement and field or false.
func (a *Accumulator) Int64Field(measurement string, field string) (int64, bool) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					v, ok := value.(int64)
					return v, ok
				}
			}
		}
	}

	return 0, false
}

// Uint64Field returns the int64 value of the given measurement and field or false.
func (a *Accumulator) Uint64Field(measurement string, field string) (uint64, bool) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					v, ok := value.(uint64)
					return v, ok
				}
			}
		}
	}

	return 0, false
}

// Int32Field returns the int32 value of the given measurement and field or false.
func (a *Accumulator) Int32Field(measurement string, field string) (int32, bool) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					v, ok := value.(int32)
					return v, ok
				}
			}
		}
	}

	return 0, false
}

// FloatField returns the float64 value of the given measurement and field or false.
func (a *Accumulator) FloatField(measurement string, field string) (float64, bool) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					v, ok := value.(float64)
					return v, ok
				}
			}
		}
	}

	return 0.0, false
}

// StringField returns the string value of the given measurement and field or false.
func (a *Accumulator) StringField(measurement string, field string) (string, bool) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					v, ok := value.(string)
					return v, ok
				}
			}
		}
	}
	return "", false
}

// BoolField returns the bool value of the given measurement and field or false.
func (a *Accumulator) BoolField(measurement string, field string) (bool, bool) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					v, ok := value.(bool)
					return v, ok
				}
			}
		}
	}

	return false, false
}
