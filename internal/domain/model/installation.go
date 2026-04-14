package model

import "time"

// Installation represents a physical lift installation linked to an EPIC device.
type Installation struct {
	ID             string
	InstallationID string // External ID used by FAIN
	DeviceSerial   string // EPCL serial number
	EPCLType       EPCLType
	BatteryType    uint8
	FirmwareVer    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// EPCLType identifies the converter model.
type EPCLType uint8

const (
	EPCLTypeUnknown  EPCLType = 0x00
	EPCL2k2_324     EPCLType = 0x0D
	EPCL3k5_648     EPCLType = 0x1D
	EPCL5k5_648     EPCLType = 0x2D
	EPCLe3_3k5_648  EPCLType = 0x3D
	EPCLe3_5k5_648  EPCLType = 0x4D
)

func (t EPCLType) String() string {
	switch t {
	case EPCL2k2_324:
		return "EPCL 2k2 324"
	case EPCL3k5_648:
		return "EPCL 3k5 648"
	case EPCL5k5_648:
		return "EPCL 5k5 648"
	case EPCLe3_3k5_648:
		return "EPCL-e3 3k5 648"
	case EPCLe3_5k5_648:
		return "EPCL-e3 5k5 648"
	default:
		return "Unknown"
	}
}

func (t EPCLType) HasLithium() bool {
	return t == EPCLe3_3k5_648 || t == EPCLe3_5k5_648
}
