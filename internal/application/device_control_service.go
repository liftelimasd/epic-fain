package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liftel/epic-fain/internal/domain/model"
	"github.com/liftel/epic-fain/internal/domain/port"
)

// DeviceControlAppService implements the DeviceControlService driving port.
type DeviceControlAppService struct {
	canSender port.CANSender
	auditRepo port.AuditLogRepository
}

func NewDeviceControlAppService(
	canSender port.CANSender,
	auditRepo port.AuditLogRepository,
) *DeviceControlAppService {
	return &DeviceControlAppService{
		canSender: canSender,
		auditRepo: auditRepo,
	}
}

func (s *DeviceControlAppService) EnableVVVF(ctx context.Context, installationID string, actor string) error {
	cmd := model.VVVFCommand{InstallationID: installationID, DCBusEnable: true}
	frame, _ := model.NewCANFrame(model.MsgIDVVVFControl, cmd.Encode())

	if err := s.canSender.Send(ctx, installationID, frame); err != nil {
		return fmt.Errorf("sending enable VVVF: %w", err)
	}

	return s.audit(ctx, installationID, model.AuditActionEnableVVVF, actor, "Enabled VVVF DC bus")
}

func (s *DeviceControlAppService) DisableVVVF(ctx context.Context, installationID string, actor string) error {
	cmd := model.VVVFCommand{InstallationID: installationID, DCBusEnable: false}
	frame, _ := model.NewCANFrame(model.MsgIDVVVFControl, cmd.Encode())

	if err := s.canSender.Send(ctx, installationID, frame); err != nil {
		return fmt.Errorf("sending disable VVVF: %w", err)
	}

	return s.audit(ctx, installationID, model.AuditActionDisableVVVF, actor, "Disabled VVVF DC bus")
}

func (s *DeviceControlAppService) ResetInverter(ctx context.Context, installationID string, actor string) error {
	cmd := model.VVVFCommand{InstallationID: installationID, DCBusEnable: true, ResetInverter: true}
	frame, _ := model.NewCANFrame(model.MsgIDVVVFControl, cmd.Encode())

	if err := s.canSender.Send(ctx, installationID, frame); err != nil {
		return fmt.Errorf("sending reset inverter: %w", err)
	}

	return s.audit(ctx, installationID, model.AuditActionResetInverter, actor, "Reset AC/DC power (20s cycle)")
}

func (s *DeviceControlAppService) EnableSuperfast(ctx context.Context, installationID string, duration time.Duration, actor string) error {
	// Superfast: change measurement period to 250ms via 0xEF01
	cmd := model.MeasureConfigCommand{
		InstallationID: installationID,
		Enable:         true,
		PeriodMs:       250,
	}
	frame, _ := model.NewCANFrame(model.MsgIDMeasureConfig, cmd.Encode())

	if err := s.canSender.Send(ctx, installationID, frame); err != nil {
		return fmt.Errorf("sending superfast config: %w", err)
	}

	detail := fmt.Sprintf("Enabled superfast capture (250ms) for %v", duration)
	return s.audit(ctx, installationID, model.AuditActionSuperfast, actor, detail)
}

func (s *DeviceControlAppService) RequestDeviceInfo(ctx context.Context, installationID string, actor string) error {
	cmd := model.InfoRequestCommand{
		InstallationID: installationID,
		RequestCode:    model.InfoRequestDeviceInfo,
	}
	frame, _ := model.NewCANFrame(model.MsgIDInfoRequest, cmd.Encode())

	if err := s.canSender.Send(ctx, installationID, frame); err != nil {
		return fmt.Errorf("sending info request: %w", err)
	}

	return s.audit(ctx, installationID, model.AuditActionCreate, actor, "Requested device info (0xEF0E)")
}

func (s *DeviceControlAppService) audit(ctx context.Context, installationID string, action model.AuditAction, actor, detail string) error {
	return s.auditRepo.Save(ctx, model.AuditLog{
		ID:             uuid.New().String(),
		InstallationID: installationID,
		Action:         action,
		Actor:          actor,
		Detail:         detail,
		Timestamp:      time.Now(),
	})
}
