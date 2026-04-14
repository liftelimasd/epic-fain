package model

import "time"

// MessageID represents a CAN message identifier.
type MessageID uint16

// Inbound messages: Equipos Epic → Plataforma
const (
	MsgIDStatus      MessageID = 0xFF00 // 7 bytes – estado general
	MsgIDMeasurement MessageID = 0xFF01 // 8 bytes – mediciones
	MsgIDCurrents    MessageID = 0xFF02 // 4 bytes – corrientes
	MsgIDLithium     MessageID = 0xFF03 // 8 bytes – datos litio (solo lithium-enabled)
	MsgIDDeviceInfo  MessageID = 0xFF0E // 8 bytes – info dispositivo
)

// Outbound messages: Plataforma → Equipos Epic
const (
	MsgIDVVVFControl   MessageID = 0xEF00 // 1 byte – control VVVF/reset
	MsgIDMeasureConfig MessageID = 0xEF01 // 3 bytes – config mediciones
	MsgIDInfoRequest   MessageID = 0xEF0E // 1 byte – solicitud info
)

// CANFrame is a raw CAN bus frame as received from the device.
type CANFrame struct {
	MessageID  MessageID
	DLC        uint8 // Data Length Code
	Data       []byte
	ReceivedAt time.Time
}

// NewCANFrame creates a validated CAN frame.
func NewCANFrame(msgID MessageID, data []byte) (CANFrame, error) {
	if len(data) > 8 {
		return CANFrame{}, ErrInvalidDLC
	}
	return CANFrame{
		MessageID:  msgID,
		DLC:        uint8(len(data)),
		Data:       data,
		ReceivedAt: time.Now(),
	}, nil
}
