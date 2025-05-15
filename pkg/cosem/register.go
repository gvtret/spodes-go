package cosem

import (
	"fmt"
	"reflect"

	"github.com/gvtret/spodes-go/pkg/axdr"
)

const (
	UnitYear               byte = 1   // a time year
	UnitMonth              byte = 2   // mo time month
	UnitWeek               byte = 3   // wk time week 7*24*60*60 s
	UnitDay                byte = 4   // d time day 24*60*60 s
	UnitHour               byte = 5   // h time hour 60*60 s
	UnitMinute             byte = 6   // min time minute 60 s
	UnitSecond             byte = 7   // s time (t) second s
	UnitAngleDegree        byte = 8   // ° (phase) angle degree rad*180/
	UnitDegreeCelsius      byte = 9   // °C temperature (T) degree-celsius K-273.15
	UnitCurrency           byte = 10  // currency (local) currency
	UnitMetre              byte = 11  // m length (l) metre m
	UnitMetrePerSecond     byte = 12  // m/s speed (v) metre per second m/s
	UnitVolume             byte = 13  // m3 volume (V)
	UnitVolumeCorrected    byte = 14  // m3 corrected volume a cubic metre m3
	UnitVolumePerHour      byte = 15  // m3/h volume flux cubic metre per hour m3/(60*60s)
	UnitVolumePerHourCorr  byte = 16  // m3/h corrected volume flux a cubic metre per hour m3/(60*60s)
	UnitVolumePerDay       byte = 17  // m3/d volume flux  m3/(24*60*60s)
	UnitVolumePerDayCorr   byte = 18  // m3/d corrected volume flux a  m3/(24*60*60s)
	UnitLitre              byte = 19  // l volume litre 10-3 m3
	UnitKilogram           byte = 20  // kg mass (m) kilogram
	UnitNewton             byte = 21  // N force (F) newton
	UnitNewtonMetre        byte = 22  // Nm energy  newton meter J = Nm = Ws
	UnitPascal             byte = 23  // Pa pressure (p) pascal N/m2
	UnitBar                byte = 24  // bar pressure (p) bar 105 N/m2
	UnitJoule              byte = 25  // J energy  joule J = Nm = Ws
	UnitJoulePerHour       byte = 26  // J/h thermal power joule per hour J/(60*60s)
	UnitWatt               byte = 27  // W active power (P) watt W = J/s
	UnitVoltAmpere         byte = 28  // VA apparent power (S) volt-ampere
	UnitVoltAmpereR        byte = 29  // var reactive power (Q) var
	UnitWattHour           byte = 30  // Wh active energy rW, active energy meter constant or pulse value watt-hour W*(60*60s)
	UnitVoltAmpereHour     byte = 31  // VAh apparent energy rS, apparent energy meter constant or pulse value volt-ampere-hour VA*(60*60s)
	UnitVoltAmpereRHour    byte = 32  // varh reactive energy rB, reactive energy meter constant or pulse value var-hour var*(60*60s)
	UnitAmpere             byte = 33  // A current (I) ampere A
	UnitCoulomb            byte = 34  // C electrical charge (Q) coulomb C = As
	UnitVolt               byte = 35  // V voltage (U) volt V
	UnitVoltPerMetre       byte = 36  // V/m electric field strength (E) volt per metre V/m
	UnitFarad              byte = 37  // F capacitance (C) farad C/V = As/V
	UnitOhm                byte = 38  // resistance (R // ohm = V/A
	UnitOhmPerMetre        byte = 39  // Ohm * m2/m resistivity   Ohm * m
	UnitWeber              byte = 40  // Wb magnetic flux weber Wb = Vs
	UnitTesla              byte = 41  // T magnetic flux density (B) tesla Wb/m2
	UnitAmperePerMetre     byte = 42  // A/m magnetic field strength (H) ampere per metre A/m
	UnitHenry              byte = 43  // H inductance (L) henry H = Wb/A
	UnitHerz               byte = 44  // Hz frequency (f, ω) hertz 1/s
	UnitWattHourPulse      byte = 45  // 1/(Wh) RW, active energy meter constant or pulse value
	UnitVoltAmpereRPulse   byte = 46  // 1/(varh) RB, reactive energy meter constant or pulse value
	UnitVoltAmperePulse    byte = 47  // 1/(VAh) RS, apparent energy meter constant or pulse value
	UnitWeekVoltSquared    byte = 48  // V2h volt-squared hour, rU2h , volt-squared hour meter constant or pulse value volt-squared-hours V2(60*60s)
	UnitWeekAmpereSquared  byte = 49  // A2h ampere-squared hour, rI2h , ampere-squared hour meter constant or pulse value ampere-squared-hours A2(60*60s)
	UnitWeekKiloPerSecond  byte = 50  // kg/s mass flux kilogram per second kg/s
	UnitSiemens            byte = 51  // S, mho conductance siemens 1/
	UnitKelvin             byte = 52  // K temperature (T) kelvin
	UnitVoltSquaredPulse   byte = 53  // 1/(V2h) RU2h, volt-squared hour meter constant or pulse value
	UnitAmpereSquaredPulse byte = 54  // 1/(A2h) RI2h, ampere-squared hour meter constant or pulse value
	UnitVolumePulse        byte = 55  // 1/m3 RV, meter constant or pulse value (volume //
	UnitPercent            byte = 56  //  percentage %
	UnitAmpereHour         byte = 57  // Ah ampere-hours Ampere-hour
	UnitJoulePerVolume     byte = 60  // Wh/m3 energy per volume 3,6*103 J/m3
	UnitWobbe              byte = 61  // J/m3 calorific value, wobbe
	UnitMole               byte = 62  // Mol % molar fraction of gas composition mole percent (Basic gas composition unit)
	UnitGramPerVolume      byte = 63  // g/m3 mass density, quantity of material  (Gas analysis, accompanying elements)
	UnitPascalSecond       byte = 64  // Pa s dynamic viscosity pascal second (Characteristic of gas stream)
	UnitJoulePerKilo       byte = 65  // J/kg specific energy NOTE The amount of energy per unit of mass of a substance Joule / kilogram m2 . kg . s -2 / kg = m2 . s –2
	UnitdBMilliWatt        byte = 70  // dBm signal strength, dB milliwatt (e.g. of GSM radio systems)
	UnitdBMicroVolt        byte = 71  // dbµV signal strength, dB microvolt
	UnitdB                 byte = 72  // dB logarithmic unit that expresses the ratio between two values of a physical quantity
	UnitReserved           byte = 253 //  reserved
	UnitOther              byte = 254 // other other unit
	UnitCount              byte = 255 // count no unit, unitless, count
)

// RegisterClassID is the class ID for the "Register" interface class as defined in IEC 62056-6-2.
const RegisterClassID uint16 = 3

// RegisterVersion is the version of the "Register" interface class.
const RegisterVersion byte = 0

// ScalerUnit represents the scaler and unit structure for the Register class.
type ScalerUnit struct {
	Scaler int8  // Scaling factor (e.g., -2 for 10^-2)
	Unit   uint8 // Unit code (e.g., 27 for Wh, per IEC 62056-6-2)
}

// Register represents the COSEM "Register" interface class (Class ID: 3, Version: 0).
// It stores a process value with a scaler and unit, identified by a logical name (OBIS code).
type Register struct {
	BaseImpl
}

// NewRegister creates a new instance of the "Register" interface class.
// Parameters:
// - obis: OBIS code for logical_name (6-byte octet-string).
// - value: Process value (any A-XDR supported type, e.g., uint32, float32).
// - scalerUnit: Scaler and unit structure (scaler as int8, unit as uint8).
func NewRegister(obis ObisCode, value interface{}, scalerUnit ScalerUnit) (*Register, error) {
	// Verify the OBIS code
	if _, err := NewObisCodeFromString(obis.String()); err != nil {
		return nil, fmt.Errorf("invalid OBIS code: %v", err)
	}

	// Verify the value type is supported by A-XDR
	if _, err := axdr.Encode(value); err != nil {
		return nil, fmt.Errorf("invalid value type for A-XDR encoding: %v", err)
	}

	// Verify the scaler_unit type is supported by A-XDR
	scalerUnitStruct := axdr.Structure{scalerUnit.Scaler, scalerUnit.Unit}
	if _, err := axdr.Encode(scalerUnitStruct); err != nil {
		return nil, fmt.Errorf("invalid scaler_unit type for A-XDR encoding: %v", err)
	}

	// Define attributes
	attributes := map[byte]AttributeDescriptor{
		1: { // logical_name
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  obis,
		},
		2: { // value
			Type:   reflect.TypeOf(value),
			Access: AttributeRead | AttributeWrite, // Configurator can write
			Value:  value,
		},
		3: { // scaler_unit
			Type:   reflect.TypeOf(scalerUnitStruct),
			Access: AttributeRead | AttributeWrite, // Configurator can write
			Value:  scalerUnitStruct,
		},
	}

	// Define methods
	methods := map[byte]MethodDescriptor{
		1: { // reset
			Access:     MethodAccessAllowed,
			ParamTypes: []reflect.Type{},
			ReturnType: nil,
			Handler: func(params []interface{}) (interface{}, error) {
				// Reset value to default (e.g., 0 or empty for the type)
				defaultValue := reflect.Zero(reflect.TypeOf(value)).Interface()
				attributes[2] = AttributeDescriptor{
					Type:   reflect.TypeOf(value),
					Access: AttributeRead | AttributeWrite,
					Value:  defaultValue,
				}
				return nil, nil
			},
		},
	}

	return &Register{
		BaseImpl: BaseImpl{
			ClassID:    RegisterClassID,
			InstanceID: obis,
			Attributes: attributes,
			Methods:    methods,
		},
	}, nil
}

// GetAttribute retrieves the value of the specified attribute.
// Supported attribute IDs:
// - 1: logical_name (ObisCode)
// - 2: value (any A-XDR supported type)
// - 3: scaler_unit (axdr.Structure{scaler, unit})
func (r *Register) GetAttribute(attributeID byte) (interface{}, error) {
	return r.BaseImpl.GetAttribute(attributeID)
}

// SetAttribute sets the value of the specified attribute.
// Supported attribute IDs:
// - 2: value (must match existing type)
// - 3: scaler_unit (must be axdr.Structure{int8, uint8})
func (r *Register) SetAttribute(attributeID byte, value interface{}) error {
	if attributeID != 2 && attributeID != 3 {
		return ErrAttributeNotSupported
	}

	// Verify A-XDR encoding compatibility
	if _, err := axdr.Encode(value); err != nil {
		return fmt.Errorf("invalid value type for A-XDR encoding: %v", err)
	}

	return r.BaseImpl.SetAttribute(attributeID, value)
}

// Invoke executes the specified method.
// Supported method IDs:
// - 1: reset (resets value to default)
func (r *Register) Invoke(methodID byte, parameters []interface{}) (interface{}, error) {
	return r.BaseImpl.Invoke(methodID, parameters)
}

// GetAttributeAccess returns the access level for the specified attribute.
func (r *Register) GetAttributeAccess(attributeID byte) AttributeAccess {
	return r.BaseImpl.GetAttributeAccess(attributeID)
}

// GetMethodAccess returns the access level for the specified method.
func (r *Register) GetMethodAccess(methodID byte) MethodAccess {
	return r.BaseImpl.GetMethodAccess(methodID)
}
