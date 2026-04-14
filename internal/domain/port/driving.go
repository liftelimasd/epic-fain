package port

import (
	"context"
	"time"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// TelemetryService is the primary driving port for telemetry operations.
type TelemetryService interface {
	// IngestFrame receives a raw CAN frame, decodes it, and stores it.
	IngestFrame(ctx context.Context, installationID string, frame model.CANFrame) error

	// GetTelemetry reads telemetry records. Read does not delete.
	GetTelemetry(ctx context.Context, installationID string, from, to time.Time) ([]model.TelemetryRecord, error)

	// GetUnconsumed returns records pending consumption by FAIN.
	GetUnconsumed(ctx context.Context, installationID string, limit int) ([]model.TelemetryRecord, error)

	// AckConsumed marks a record as consumed by the FAIN consumer.
	AckConsumed(ctx context.Context, recordID string, actor string) error

	// DeleteConsumed removes ACK'd records. Audited.
	DeleteConsumed(ctx context.Context, ids []string, actor string) error
}

// DeviceControlService manages remote commands to EPIC devices.
type DeviceControlService interface {
	// EnableVVVF enables the VVVF DC bus.
	EnableVVVF(ctx context.Context, installationID string, actor string) error

	// DisableVVVF disables the VVVF DC bus.
	DisableVVVF(ctx context.Context, installationID string, actor string) error

	// ResetInverter triggers a 20s reset cycle of AC/DC power.
	ResetInverter(ctx context.Context, installationID string, actor string) error

	// EnableSuperfast activates superfast capture mode (250ms) for a duration.
	// Cannot be reactivated while already active.
	EnableSuperfast(ctx context.Context, installationID string, duration time.Duration, actor string) error

	// RequestDeviceInfo sends a 0xEF0E to get device info.
	RequestDeviceInfo(ctx context.Context, installationID string, actor string) error
}

// AlertService manages alert evaluation and delivery.
type AlertService interface {
	// EvaluateFrame checks a decoded telemetry record against active rules.
	EvaluateFrame(ctx context.Context, record model.TelemetryRecord) ([]model.Alert, error)

	// GetAlerts returns alerts for an installation in a time range.
	GetAlerts(ctx context.Context, installationID string, from, to time.Time) ([]model.Alert, error)

	// AcknowledgeAlert marks an alert as acknowledged.
	AcknowledgeAlert(ctx context.Context, alertID string) error
}

// InstallationService manages installation/device registration.
type InstallationService interface {
	Register(ctx context.Context, inst model.Installation) error
	Get(ctx context.Context, installationID string) (*model.Installation, error)
	List(ctx context.Context) ([]model.Installation, error)
	UpdateDeviceInfo(ctx context.Context, installationID string, info model.DeviceInfoPayload) error
}
