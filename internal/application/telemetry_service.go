package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liftel/epic-fain/internal/domain/model"
	"github.com/liftel/epic-fain/internal/domain/port"
	"github.com/liftel/epic-fain/internal/domain/service"
)

// TelemetryAppService implements the TelemetryService driving port.
type TelemetryAppService struct {
	telemetryRepo port.TelemetryRepository
	auditRepo     port.AuditLogRepository
	alertSvc      port.AlertService
	mqttPub       port.MQTTPublisher
	decoder       *service.CANDecoder
}

func NewTelemetryAppService(
	telemetryRepo port.TelemetryRepository,
	auditRepo port.AuditLogRepository,
	alertSvc port.AlertService,
	mqttPub port.MQTTPublisher,
	decoder *service.CANDecoder,
) *TelemetryAppService {
	return &TelemetryAppService{
		telemetryRepo: telemetryRepo,
		auditRepo:     auditRepo,
		alertSvc:      alertSvc,
		mqttPub:       mqttPub,
		decoder:       decoder,
	}
}

// IngestFrame decodes a CAN frame, stores it, evaluates alert rules, and publishes alerts.
func (s *TelemetryAppService) IngestFrame(ctx context.Context, installationID string, frame model.CANFrame) error {
	decoded, err := s.decoder.Decode(frame)
	if err != nil {
		return fmt.Errorf("decoding CAN frame: %w", err)
	}

	record := model.TelemetryRecord{
		ID:             uuid.New().String(),
		InstallationID: installationID,
		MessageID:      frame.MessageID,
		RawData:        frame.Data,
		DecodedPayload: decoded,
		CapturedAt:     frame.ReceivedAt,
		ReceivedAt:     time.Now(),
	}

	if err := s.telemetryRepo.Store(ctx, record); err != nil {
		return fmt.Errorf("storing telemetry record: %w", err)
	}

	// Audit the creation
	audit := model.AuditLog{
		ID:             uuid.New().String(),
		InstallationID: installationID,
		Action:         model.AuditActionCreate,
		Actor:          "system",
		Detail:         fmt.Sprintf("Ingested CAN frame 0x%04X (%d bytes)", uint16(frame.MessageID), frame.DLC),
		Timestamp:      time.Now(),
	}
	_ = s.auditRepo.Save(ctx, audit) // Best effort for ingestion audit

	// Evaluate alert rules against the new record
	if s.alertSvc != nil {
		alerts, err := s.alertSvc.EvaluateFrame(ctx, record)
		if err == nil {
			for _, alert := range alerts {
				if s.mqttPub != nil {
					_ = s.mqttPub.PublishAlert(ctx, installationID, alert)
				}
			}
		}
	}

	return nil
}

func (s *TelemetryAppService) GetTelemetry(ctx context.Context, installationID string, from, to time.Time) ([]model.TelemetryRecord, error) {
	return s.telemetryRepo.FindByInstallation(ctx, installationID, from, to)
}

func (s *TelemetryAppService) GetUnconsumed(ctx context.Context, installationID string, limit int) ([]model.TelemetryRecord, error) {
	return s.telemetryRepo.FindUnconsumed(ctx, installationID, limit)
}

// AckConsumed marks a record as consumed. Audited.
func (s *TelemetryAppService) AckConsumed(ctx context.Context, recordID string, actor string) error {
	rec, err := s.telemetryRepo.FindByID(ctx, recordID)
	if err != nil {
		return err
	}

	if err := s.telemetryRepo.AckConsumed(ctx, recordID); err != nil {
		return err
	}

	return s.auditRepo.Save(ctx, model.AuditLog{
		ID:             uuid.New().String(),
		InstallationID: rec.InstallationID,
		Action:         model.AuditActionACK,
		Actor:          actor,
		Detail:         fmt.Sprintf("ACK telemetry record %s", recordID),
		Timestamp:      time.Now(),
	})
}

// DeleteConsumed removes ACK'd records. Audited. Only consumed records can be deleted.
func (s *TelemetryAppService) DeleteConsumed(ctx context.Context, ids []string, actor string) error {
	if err := s.telemetryRepo.DeleteConsumed(ctx, ids); err != nil {
		return err
	}

	// Audit the deletion (grouped)
	return s.auditRepo.Save(ctx, model.AuditLog{
		ID:             uuid.New().String(),
		InstallationID: "batch",
		Action:         model.AuditActionDelete,
		Actor:          actor,
		Detail:         fmt.Sprintf("Deleted %d consumed telemetry records", len(ids)),
		Timestamp:      time.Now(),
	})
}
