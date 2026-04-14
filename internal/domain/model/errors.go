package model

import "errors"

var (
	ErrInvalidDLC          = errors.New("CAN frame DLC exceeds 8 bytes")
	ErrUnknownMessageID    = errors.New("unknown CAN message ID")
	ErrInvalidPayload      = errors.New("payload does not match expected DLC for message type")
	ErrInstallationNotFound = errors.New("installation not found")
	ErrTelemetryNotFound   = errors.New("telemetry record not found")
	ErrAlreadyConsumed     = errors.New("telemetry record already consumed")
	ErrSuperfastActive     = errors.New("superfast mode already active")
	ErrInvalidRetention    = errors.New("retention period must be positive")
	ErrUnauthorized        = errors.New("invalid or missing API key")
)
