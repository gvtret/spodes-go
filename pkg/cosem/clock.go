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

	methods := map[byte]MethodDescriptor{}

	return &Clock{
		BaseImpl: BaseImpl{
			ClassID:    ClockClassID,
			InstanceID: obis,
			Attributes: attributes,
			Methods:    methods,
		},
	}, nil
}
