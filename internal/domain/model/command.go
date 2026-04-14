package model

import "time"

// CaptureMode defines the data acquisition rate.
type CaptureMode string

const (
	CaptureModeNormal    CaptureMode = "NORMAL"    // 1s period
	CaptureModeSuperfast CaptureMode = "SUPERFAST"  // 250ms internal capture
)

// CaptureState tracks the current capture configuration for an installation.
type CaptureState struct {
	InstallationID string
	Mode           CaptureMode
	Active         bool
	SuperfastEnd   *time.Time // When superfast expires (nil if normal)
	MeasurePeriodMs uint16    // Current measurement period in ms
}

// VVVFCommand represents a command to send to EPIC via 0xEF00.
type VVVFCommand struct {
	InstallationID string
	DCBusEnable    bool // Bit 0: enable VVVF DC bus
	ResetInverter  bool // Bit 1: reset AC/DC power (20s cycle)
}

// Encode serializes the command to a 1-byte CAN payload.
func (c VVVFCommand) Encode() []byte {
	var b byte
	if c.DCBusEnable {
		b |= 0x01
	}
	if c.ResetInverter {
		b |= 0x02
	}
	return []byte{b}
}

// MeasureConfigCommand represents 0xEF01 configuration.
type MeasureConfigCommand struct {
	InstallationID string
	Enable         bool   // Bit 0: enable measurement messages
	PeriodMs       uint16 // 50-1000 ms
}

// Encode serializes to 3-byte CAN payload.
func (c MeasureConfigCommand) Encode() []byte {
	data := make([]byte, 3)
	if c.Enable {
		data[0] = 0x01
	}
	data[1] = byte(c.PeriodMs >> 8)
	data[2] = byte(c.PeriodMs & 0xFF)
	return data
}

// InfoRequestCommand represents 0xEF0E.
type InfoRequestCommand struct {
	InstallationID string
	RequestCode    uint8 // 51=device info (0xFF0E), 54=measurements (0xFF01+0xFF03)
}

// Encode serializes to 1-byte CAN payload.
func (c InfoRequestCommand) Encode() []byte {
	return []byte{c.RequestCode}
}

const (
	InfoRequestDeviceInfo    uint8 = 51 // Decimal, triggers 0xFF0E response
	InfoRequestMeasurements  uint8 = 54 // Decimal, triggers 0xFF01 + 0xFF03
)
