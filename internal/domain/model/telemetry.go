package model

import "time"

// TelemetryRecord stores a decoded CAN message associated to an installation.
// This is the central entity persisted in the temporal DB.
type TelemetryRecord struct {
	ID             string
	InstallationID string
	MessageID      MessageID
	RawData        []byte
	DecodedPayload interface{} // One of the typed payloads below
	CapturedAt     time.Time   // When the device produced it
	ReceivedAt     time.Time   // When the platform received it
	ConsumedAt     *time.Time  // Nil until ACK'd by consumer (FAIN)
}

// IsConsumed returns true when the record has been acknowledged.
func (t TelemetryRecord) IsConsumed() bool {
	return t.ConsumedAt != nil
}

// -----------------------------------------------------------------
// Value Objects: decoded payloads per message type
// -----------------------------------------------------------------

// VVVFStatus represents VVVF operating state.
type VVVFStatus uint8

const (
	VVVFOff                 VVVFStatus = 0
	VVVFOnNoOvercharge      VVVFStatus = 1
	VVVFOnOvercharge        VVVFStatus = 2
	VVVFOnOverchargeWarning VVVFStatus = 3 // <2s to reduce
)

// StatusPayload is decoded from 0xFF00 (7 bytes).
type StatusPayload struct {
	VVVF               VVVFStatus
	HardwareEnableEcho bool  // Pin 3A-3B
	SystemStatus       bool  // Pin 3A-3C
	ChargerStatus      bool  // Batteries being charged
	BatterySoC         uint8 // 0-100 %
	WarningCode        WarningCode
	ErrorCode          ErrorCode
	ErrorValue         int16
	ACInverterEnabled  bool
	ACChargerEnabled   bool
}

// WarningCode enumerates EPCL warnings.
type WarningCode uint8

const (
	WarningNone        WarningCode = 0
	WarningTemperature WarningCode = 1
	WarningEarthFault  WarningCode = 2
)

func (w WarningCode) String() string {
	switch w {
	case WarningNone:
		return "No warning"
	case WarningTemperature:
		return "Temperature warning"
	case WarningEarthFault:
		return "Earth fault detected"
	default:
		return "Unknown"
	}
}

// ErrorCode enumerates EPCL errors.
type ErrorCode uint8

const (
	ErrorNone                  ErrorCode = 0
	ErrorDCLinkVoltageSW       ErrorCode = 1
	ErrorBatteryVoltageSW      ErrorCode = 2
	ErrorDCLinkOvercurrentSW   ErrorCode = 3
	ErrorHardwareInternal      ErrorCode = 4
	ErrorOvertemperature       ErrorCode = 5
	ErrorDCLinkSwitchOn        ErrorCode = 6
	ErrorEarthFaultDCLink      ErrorCode = 7
	ErrorSoftwareInternal2     ErrorCode = 8
	ErrorReserved              ErrorCode = 9
	ErrorBatteryVoltageHW      ErrorCode = 10
	ErrorDCLinkOvercurrentHW   ErrorCode = 11
)

func (e ErrorCode) String() string {
	names := map[ErrorCode]string{
		ErrorNone:                "No error",
		ErrorDCLinkVoltageSW:    "DC link voltage error (SW)",
		ErrorBatteryVoltageSW:   "Battery voltage error (SW)",
		ErrorDCLinkOvercurrentSW: "DC link overcurrent (SW)",
		ErrorHardwareInternal:   "Hardware internal error",
		ErrorOvertemperature:    "Overtemperature (>80°C)",
		ErrorDCLinkSwitchOn:     "DC link switch-on error",
		ErrorEarthFaultDCLink:   "Earth fault in DC link",
		ErrorSoftwareInternal2:  "Software internal error 2",
		ErrorReserved:           "Reserved",
		ErrorBatteryVoltageHW:   "Battery voltage error (HW)",
		ErrorDCLinkOvercurrentHW: "DC link overcurrent (HW)",
	}
	if n, ok := names[e]; ok {
		return n
	}
	return "Unknown"
}

// ErrorValueUnit returns the unit for the error value given the error code.
func (e ErrorCode) ErrorValueUnit() string {
	switch e {
	case ErrorDCLinkVoltageSW, ErrorDCLinkSwitchOn:
		return "V (1V/1)"
	case ErrorBatteryVoltageSW, ErrorBatteryVoltageHW:
		return "V (0.1V/1)"
	case ErrorDCLinkOvercurrentSW, ErrorDCLinkOvercurrentHW:
		return "A (0.1A/1)"
	case ErrorOvertemperature:
		return "°C"
	default:
		return ""
	}
}

// MeasurementPayload is decoded from 0xFF01 (8 bytes).
type MeasurementPayload struct {
	BatteryVoltage float64 // V (scaling 0.1V/1)
	BatteryCurrent int16   // A (1A/1, positive=charge, negative=discharge)
	DCLinkPower    int16   // W (1W/1, negative=consume, positive=regenerate)
	DCLinkVoltage  uint16  // V (1V/1)
}

// CurrentsPayload is decoded from 0xFF02 (4 bytes).
type CurrentsPayload struct {
	DCLinkCurrent      float64 // A (0.1A/1, signed)
	ACChargerCurrent   float64 // A (0.1A/1, unsigned)
	ACInverterCurrent  float64 // A (0.1A/1, unsigned)
	SolarChargerCurrent float64 // A (0.1A/1, unsigned)
}

// LithiumPayload is decoded from 0xFF03 (8 bytes). Only lithium-enabled EPCLs.
type LithiumPayload struct {
	SoH              uint8  // State of Health %
	BMSStatus        uint16 // Internal BMS code
	LithiumTemp      int8   // °C
	CycleCount       uint16
	PackVoltage      uint16 // Internal pack voltage
}

// DeviceInfoPayload is decoded from 0xFF0E (8 bytes).
type DeviceInfoPayload struct {
	EPCLType       EPCLType
	BatteryType    uint8
	FirmwareVersion string // Decoded from 20-bit field
	SerialNumber   uint32
}
