package cosem

import (
	"reflect"
	"time"
)

// ClockClassID is the class ID for the "Clock" interface class.
const ClockClassID uint16 = 8

// ClockVersion is the version of the "Clock" interface class.
const ClockVersion byte = 0

// Clock represents the COSEM "Clock" interface class.
type Clock struct {
	BaseImpl

	presetAdjustTime *time.Time
}

// NewClock creates a new instance of the "Clock" interface class.
func NewClock(obis ObisCode) (*Clock, error) {
	attributes := map[byte]AttributeDescriptor{
		1: { // logical_name
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  obis,
		},
		2: { // time
			Type:   reflect.TypeOf(time.Now()),
			Access: AttributeRead | AttributeWrite,
			Value:  time.Now(),
		},
		3: { // time_zone
			Type:   reflect.TypeOf(int16(0)),
			Access: AttributeRead | AttributeWrite,
			Value:  int16(0),
		},
		4: { // status
			Type:   reflect.TypeOf(uint8(0)),
			Access: AttributeRead | AttributeWrite,
			Value:  uint8(0),
		},
		5: { // daylight_savings_begin
			Type:   reflect.TypeOf(time.Time{}),
			Access: AttributeRead | AttributeWrite,
			Value:  time.Time{},
		},
		6: { // daylight_savings_end
			Type:   reflect.TypeOf(time.Time{}),
			Access: AttributeRead | AttributeWrite,
			Value:  time.Time{},
		},
		7: { // daylight_savings_deviation
			Type:   reflect.TypeOf(int8(0)),
			Access: AttributeRead | AttributeWrite,
			Value:  int8(0),
		},
		8: { // daylight_savings_enabled
			Type:   reflect.TypeOf(false),
			Access: AttributeRead | AttributeWrite,
			Value:  false,
		},
		9: { // clock_base
			Type:   reflect.TypeOf(uint8(0)),
			Access: AttributeRead,
			Value:  uint8(0),
		},
	}

	clock := &Clock{}
	clock.BaseImpl = BaseImpl{
		ClassID:    ClockClassID,
		InstanceID: obis,
		Attributes: attributes,
		Methods:    make(map[byte]MethodDescriptor),
	}

	clock.Methods[1] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: nil,
		Handler:    clock.adjustToQuarterHour,
	}
	clock.Methods[2] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{reflect.TypeOf(uint16(0))},
		Handler:    clock.adjustToMeasuringPeriod,
	}
	clock.Methods[3] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: nil,
		Handler:    clock.adjustToMinute,
	}
	clock.Methods[4] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{reflect.TypeOf(time.Time{})},
		Handler:    clock.presetAdjustingTime,
	}
	clock.Methods[5] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: nil,
		Handler:    clock.adjustToPresetTime,
	}
	clock.Methods[6] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{reflect.TypeOf(int32(0))},
		Handler:    clock.shiftTime,
	}

	return clock, nil
}

const (
	clockStatusInvalid  = 0x01
	clockStatusDoubtful = 0x02
)

func (c *Clock) adjustToQuarterHour(_ []interface{}) (interface{}, error) {
	current, err := c.currentTime()
	if err != nil {
		return nil, err
	}

	base := current.Truncate(15 * time.Minute)
	if current.Sub(base) >= (15 * time.Minute / 2) {
		base = base.Add(15 * time.Minute)
	}
	base = base.Truncate(time.Minute)

	c.setTime(base)
	c.clearStatusFlags(clockStatusInvalid)
	return nil, nil
}

func (c *Clock) adjustToMeasuringPeriod(params []interface{}) (interface{}, error) {
	period := params[0].(uint16)
	if period == 0 {
		return nil, ErrInvalidParameter
	}

	current, err := c.currentTime()
	if err != nil {
		return nil, err
	}

	duration := time.Duration(period) * time.Minute
	startOfDay := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())
	elapsed := current.Sub(startOfDay)
	aligned := startOfDay.Add(elapsed - (elapsed % duration))

	c.setTime(aligned)
	c.clearStatusFlags(clockStatusInvalid)
	return nil, nil
}

func (c *Clock) adjustToMinute(_ []interface{}) (interface{}, error) {
	current, err := c.currentTime()
	if err != nil {
		return nil, err
	}

	truncated := current.Truncate(time.Minute)
	if current.Sub(truncated) >= 30*time.Second {
		truncated = truncated.Add(time.Minute)
	}

	c.setTime(truncated)
	c.clearStatusFlags(clockStatusInvalid)
	return nil, nil
}

func (c *Clock) presetAdjustingTime(params []interface{}) (interface{}, error) {
	preset := params[0].(time.Time)
	presetCopy := preset
	c.presetAdjustTime = &presetCopy
	c.setStatusFlags(clockStatusDoubtful)
	return nil, nil
}

func (c *Clock) adjustToPresetTime(_ []interface{}) (interface{}, error) {
	if c.presetAdjustTime == nil {
		return nil, ErrInvalidParameter
	}

	c.setTime(*c.presetAdjustTime)
	c.presetAdjustTime = nil
	c.clearStatusFlags(clockStatusInvalid | clockStatusDoubtful)
	return nil, nil
}

func (c *Clock) shiftTime(params []interface{}) (interface{}, error) {
	delta := time.Duration(params[0].(int32)) * time.Second

	current, err := c.currentTime()
	if err != nil {
		return nil, err
	}

	c.setTime(current.Add(delta))
	c.clearStatusFlags(clockStatusInvalid)
	return nil, nil
}

func (c *Clock) currentTime() (time.Time, error) {
	attr, ok := c.Attributes[2]
	if !ok {
		return time.Time{}, ErrAttributeNotSupported
	}
	current, ok := attr.Value.(time.Time)
	if !ok {
		return time.Time{}, ErrInvalidValueType
	}
	return current, nil
}

func (c *Clock) setTime(t time.Time) {
	attr := c.Attributes[2]
	attr.Value = t
	c.Attributes[2] = attr
}

func (c *Clock) currentStatus() uint8 {
	attr := c.Attributes[4]
	status, _ := attr.Value.(uint8)
	return status
}

func (c *Clock) setStatus(status uint8) {
	attr := c.Attributes[4]
	attr.Value = status
	c.Attributes[4] = attr
}

func (c *Clock) clearStatusFlags(flags uint8) {
	if flags == 0 {
		return
	}
	status := c.currentStatus() &^ flags
	c.setStatus(status)
}

func (c *Clock) setStatusFlags(flags uint8) {
	if flags == 0 {
		return
	}
	status := c.currentStatus() | flags
	c.setStatus(status)
}
