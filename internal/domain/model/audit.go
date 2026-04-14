package model

import "time"

// AuditAction categorizes the type of audited operation.
type AuditAction string

const (
	AuditActionCreate        AuditAction = "CREATE"
	AuditActionDelete        AuditAction = "DELETE"
	AuditActionACK           AuditAction = "ACK"
	AuditActionRestart       AuditAction = "RESTART"
	AuditActionEnableVVVF    AuditAction = "ENABLE_VVVF"
	AuditActionDisableVVVF   AuditAction = "DISABLE_VVVF"
	AuditActionSuperfast     AuditAction = "SUPERFAST"
	AuditActionResetInverter AuditAction = "RESET_INVERTER"
)

// AuditLog records every remote action and data operation for traceability.
type AuditLog struct {
	ID             string
	InstallationID string
	Action         AuditAction
	Actor          string // Who: API key identifier or system
	Detail         string // What specifically
	Timestamp      time.Time
}
