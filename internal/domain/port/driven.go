package port

import (
	"context"
	"time"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// MQTTPublisher publishes alerts to MQTT subtopics per installation.
type MQTTPublisher interface {
	PublishAlert(ctx context.Context, installationID string, alert model.Alert) error
}

// CANSender sends commands to EPIC devices via CAN bus.
type CANSender interface {
	Send(ctx context.Context, installationID string, frame model.CANFrame) error
}

// Clock abstracts time for testability.
type Clock interface {
	Now() time.Time
}
