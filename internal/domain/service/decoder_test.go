package service

import (
	"testing"

	"github.com/liftel/epic-fain/internal/domain/model"
)

func TestDecodeStatus(t *testing.T) {
	decoder := NewCANDecoder()

	// Simulated 0xFF00: VVVF ON (1), HW echo=1, system=1, charger=0, SoC=60%, warning=0, error=0, errorVal=0, inverter=1, charger=1
	data := []byte{0x0D, 0x3C, 0x00, 0x00, 0x00, 0x00, 0x03}
	frame := model.CANFrame{MessageID: model.MsgIDStatus, DLC: 7, Data: data}

	result, err := decoder.Decode(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, ok := result.(model.StatusPayload)
	if !ok {
		t.Fatal("expected StatusPayload")
	}

	if status.VVVF != model.VVVFOnNoOvercharge {
		t.Errorf("VVVF = %d, want %d", status.VVVF, model.VVVFOnNoOvercharge)
	}
	if !status.HardwareEnableEcho {
		t.Error("expected HardwareEnableEcho = true")
	}
	if !status.SystemStatus {
		t.Error("expected SystemStatus = true")
	}
	if status.ChargerStatus {
		t.Error("expected ChargerStatus = false")
	}
	if status.BatterySoC != 60 {
		t.Errorf("BatterySoC = %d, want 60", status.BatterySoC)
	}
	if !status.ACInverterEnabled || !status.ACChargerEnabled {
		t.Error("expected both AC inverter and charger enabled")
	}
}

func TestDecodeMeasurement(t *testing.T) {
	decoder := NewCANDecoder()

	// Battery voltage: 0x0224 = 548 → 54.8V
	// Battery current: 0x0002 = 2A
	// DC link power: 0xFF00 = -256W (signed)
	// DC link voltage: 0x025D = 605V
	data := []byte{0x24, 0x02, 0x02, 0x00, 0x00, 0xFF, 0x5D, 0x02}
	frame := model.CANFrame{MessageID: model.MsgIDMeasurement, DLC: 8, Data: data}

	result, err := decoder.Decode(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := result.(model.MeasurementPayload)
	if !ok {
		t.Fatal("expected MeasurementPayload")
	}

	if m.BatteryVoltage < 54.7 || m.BatteryVoltage > 54.9 {
		t.Errorf("BatteryVoltage = %.1f, want ~54.8", m.BatteryVoltage)
	}
	if m.BatteryCurrent != 2 {
		t.Errorf("BatteryCurrent = %d, want 2", m.BatteryCurrent)
	}
	if m.DCLinkVoltage != 605 {
		t.Errorf("DCLinkVoltage = %d, want 605", m.DCLinkVoltage)
	}
}

func TestDecodeCurrents(t *testing.T) {
	decoder := NewCANDecoder()

	// DC link: 0x08 = 0.8A, charger: 0x00, inverter: 0x01 = 0.1A, solar: 0x00
	data := []byte{0x08, 0x00, 0x01, 0x00}
	frame := model.CANFrame{MessageID: model.MsgIDCurrents, DLC: 4, Data: data}

	result, err := decoder.Decode(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, ok := result.(model.CurrentsPayload)
	if !ok {
		t.Fatal("expected CurrentsPayload")
	}

	if c.DCLinkCurrent < 0.79 || c.DCLinkCurrent > 0.81 {
		t.Errorf("DCLinkCurrent = %.1f, want ~0.8", c.DCLinkCurrent)
	}
}

func TestDecodeInvalidDLC(t *testing.T) {
	decoder := NewCANDecoder()

	frame := model.CANFrame{MessageID: model.MsgIDStatus, DLC: 3, Data: []byte{0x00, 0x00, 0x00}}
	_, err := decoder.Decode(frame)
	if err != model.ErrInvalidPayload {
		t.Errorf("expected ErrInvalidPayload, got: %v", err)
	}
}
