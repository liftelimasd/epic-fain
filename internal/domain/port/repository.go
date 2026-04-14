package port

import (
	"context"
	"time"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// InstallationRepository manages installation persistence.
type InstallationRepository interface {
	Save(ctx context.Context, inst model.Installation) error
	FindByID(ctx context.Context, id string) (*model.Installation, error)
	FindByInstallationID(ctx context.Context, installationID string) (*model.Installation, error)
	FindByDeviceSerial(ctx context.Context, serial string) (*model.Installation, error)
	FindAll(ctx context.Context) ([]model.Installation, error)
	Update(ctx context.Context, inst model.Installation) error
}

// TelemetryRepository manages the temporal telemetry buffer.
// Rules: read does not delete, delete by explicit ACK only, no updates on raw data.
type TelemetryRepository interface {
	// Store persists a new telemetry record. Audit on create is mandatory.
	Store(ctx context.Context, record model.TelemetryRecord) error

	// FindByID retrieves a single record. Read does not modify state.
	FindByID(ctx context.Context, id string) (*model.TelemetryRecord, error)

	// FindByInstallation returns records for a given installation within a time range.
	FindByInstallation(ctx context.Context, installationID string, from, to time.Time) ([]model.TelemetryRecord, error)

	// FindUnconsumed returns records not yet ACK'd by the consumer.
	FindUnconsumed(ctx context.Context, installationID string, limit int) ([]model.TelemetryRecord, error)

	// AckConsumed marks a record as consumed. This is the only way to "modify" a record.
	AckConsumed(ctx context.Context, id string) error

	// DeleteConsumed removes records that have been ACK'd. Audit on delete is mandatory.
	DeleteConsumed(ctx context.Context, ids []string) error

	// PurgeExpired removes records older than the retention window.
	PurgeExpired(ctx context.Context, retention time.Duration) (int64, error)
}

// AlertRepository persists alerts.
type AlertRepository interface {
	Save(ctx context.Context, alert model.Alert) error
	FindByID(ctx context.Context, id string) (*model.Alert, error)
	FindByInstallation(ctx context.Context, installationID string, from, to time.Time) ([]model.Alert, error)
	FindUnacknowledged(ctx context.Context, installationID string) ([]model.Alert, error)
	Acknowledge(ctx context.Context, id string) error
}

// AlertRuleRepository manages alert rule definitions.
type AlertRuleRepository interface {
	Save(ctx context.Context, rule model.AlertRule) error
	FindByID(ctx context.Context, id string) (*model.AlertRule, error)
	FindEnabled(ctx context.Context) ([]model.AlertRule, error)
	Update(ctx context.Context, rule model.AlertRule) error
}

// AuditLogRepository persists audit trail entries.
type AuditLogRepository interface {
	Save(ctx context.Context, entry model.AuditLog) error
	FindByInstallation(ctx context.Context, installationID string, from, to time.Time) ([]model.AuditLog, error)
}
