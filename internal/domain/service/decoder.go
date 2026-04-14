package service

import (
	"encoding/binary"
	"fmt"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// CANDecoder decodes raw CAN frame payloads into typed domain objects.
type CANDecoder struct{}

func NewCANDecoder() *CANDecoder {
	return &CANDecoder{}
}

// Decode routes a CAN frame to the appropriate typed decoder.
func (d *CANDecoder) Decode(frame model.CANFrame) (interface{}, error) {
	switch frame.MessageID {
	case model.MsgIDStatus:
		return d.decodeStatus(frame.Data)
	case model.MsgIDMeasurement:
		return d.decodeMeasurement(frame.Data)
	case model.MsgIDCurrents:
		return d.decodeCurrents(frame.Data)
	case model.MsgIDLithium:
		return d.decodeLithium(frame.Data)
	case model.MsgIDDeviceInfo:
		return d.decodeDeviceInfo(frame.Data)
	default:
		return nil, fmt.Errorf("%w: 0x%04X", model.ErrUnknownMessageID, uint16(frame.MessageID))
	}
}

// decodeStatus decodes 0xFF00 (7 bytes).
// Bit layout:
//
//	Byte 0: [bit0-1] VVVF status, [bit2] HW enable echo, [bit3] system status, [bit4] charger status
//	Byte 1: Battery SoC (0-100%)
//	Byte 2: Warning codes
//	Byte 3: Error codes
//	Byte 4-5: Error value (int16, little-endian)
//	Byte 6: [bit0] AC inverter, [bit1] AC charger
func (d *CANDecoder) decodeStatus(data []byte) (model.StatusPayload, error) {
	if len(data) < 7 {
		return model.StatusPayload{}, model.ErrInvalidPayload
	}

	b0 := data[0]
	return model.StatusPayload{
		VVVF:               model.VVVFStatus(b0 & 0x03),
		HardwareEnableEcho: (b0>>2)&0x01 == 1,
		SystemStatus:       (b0>>3)&0x01 == 1,
		ChargerStatus:      (b0>>4)&0x01 == 1,
		BatterySoC:         data[1],
		WarningCode:        model.WarningCode(data[2]),
		ErrorCode:          model.ErrorCode(data[3]),
		ErrorValue:         int16(binary.LittleEndian.Uint16(data[4:6])),
		ACInverterEnabled:  (data[6]>>0)&0x01 == 1,
		ACChargerEnabled:   (data[6]>>1)&0x01 == 1,
	}, nil
}

// decodeMeasurement decodes 0xFF01 (8 bytes).
// Byte layout (all little-endian uint16/int16):
//
//	Byte 0-1: Battery voltage (unsigned, scaling 0.1V)
//	Byte 2-3: Battery current (signed, 1A)
//	Byte 4-5: DC link power (signed, 1W)
//	Byte 6-7: DC link voltage (unsigned, 1V)
func (d *CANDecoder) decodeMeasurement(data []byte) (model.MeasurementPayload, error) {
	if len(data) < 8 {
		return model.MeasurementPayload{}, model.ErrInvalidPayload
	}

	return model.MeasurementPayload{
		BatteryVoltage: float64(binary.LittleEndian.Uint16(data[0:2])) * 0.1,
		BatteryCurrent: int16(binary.LittleEndian.Uint16(data[2:4])),
		DCLinkPower:    int16(binary.LittleEndian.Uint16(data[4:6])),
		DCLinkVoltage:  binary.LittleEndian.Uint16(data[6:8]),
	}, nil
}

// decodeCurrents decodes 0xFF02 (4 bytes).
// Each byte is a single value:
//
//	Byte 0: DC link current (signed int8, scaling 0.1A)
//	Byte 1: AC charger current (unsigned, scaling 0.1A)
//	Byte 2: AC inverter current (unsigned, scaling 0.1A)
//	Byte 3: Solar charger current (unsigned, scaling 0.1A)
func (d *CANDecoder) decodeCurrents(data []byte) (model.CurrentsPayload, error) {
	if len(data) < 4 {
		return model.CurrentsPayload{}, model.ErrInvalidPayload
	}

	return model.CurrentsPayload{
		DCLinkCurrent:       float64(int8(data[0])) * 0.1,
		ACChargerCurrent:    float64(data[1]) * 0.1,
		ACInverterCurrent:   float64(data[2]) * 0.1,
		SolarChargerCurrent: float64(data[3]) * 0.1,
	}, nil
}

// decodeLithium decodes 0xFF03 (8 bytes). Only lithium-enabled EPCLs.
//
//	Byte 0: SoH (unsigned)
//	Byte 1-2: BMS Status (unsigned, big-endian)
//	Byte 3: Lithium Temperature (signed int8)
//	Byte 4-5: Cycle Count (unsigned, big-endian)
//	Byte 6-7: Pack Voltage (unsigned, big-endian)
func (d *CANDecoder) decodeLithium(data []byte) (model.LithiumPayload, error) {
	if len(data) < 8 {
		return model.LithiumPayload{}, model.ErrInvalidPayload
	}

	return model.LithiumPayload{
		SoH:         data[0],
		BMSStatus:   binary.BigEndian.Uint16(data[1:3]),
		LithiumTemp: int8(data[3]),
		CycleCount:  binary.BigEndian.Uint16(data[4:6]),
		PackVoltage: binary.BigEndian.Uint16(data[6:8]),
	}, nil
}

// decodeDeviceInfo decodes 0xFF0E (8 bytes).
//
//	Byte 0: EPCL type
//	Byte 1 bits 0-3: Battery type
//	Byte 1 bits 4-7 + Byte 2 + Byte 3 bits 0-3: Firmware version (20 bits)
//	Byte 4-7: Serial number (uint32, big-endian)
func (d *CANDecoder) decodeDeviceInfo(data []byte) (model.DeviceInfoPayload, error) {
	if len(data) < 8 {
		return model.DeviceInfoPayload{}, model.ErrInvalidPayload
	}

	battType := data[1] & 0x0F

	// Firmware: 20-bit field from bit 12 to bit 31
	fwRaw := uint32(data[1]>>4) | uint32(data[2])<<4 | uint32(data[3]&0x0F)<<12
	fwMajor := (fwRaw >> 16) & 0x0F
	fwMinor := (fwRaw >> 8) & 0xFF
	fwPatch := fwRaw & 0xFF
	firmware := fmt.Sprintf("%d.%d.%d", fwMajor, fwMinor, fwPatch)

	serial := binary.BigEndian.Uint32(data[4:8])

	return model.DeviceInfoPayload{
		EPCLType:        model.EPCLType(data[0]),
		BatteryType:     battType,
		FirmwareVersion: firmware,
		SerialNumber:    serial,
	}, nil
}
