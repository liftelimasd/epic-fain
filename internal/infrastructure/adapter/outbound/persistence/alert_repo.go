package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// AlertRepo implements port.AlertRepository with PostgreSQL.
type AlertRepo struct {
	db *sql.DB
}

func NewAlertRepo(db *sql.DB) *AlertRepo {
	return &AlertRepo{db: db}
}

func (r *AlertRepo) Save(ctx context.Context, alert model.Alert) error {
	payload, err := json.Marshal(alert.Payload)
	if err != nil {
		return fmt.Errorf("marshaling alert payload: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO alerts (id, installation_id, rule_id, severity, code, message, payload, triggered_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		alert.ID, alert.InstallationID, alert.RuleID,
		string(alert.Severity), alert.Code, alert.Message,
		payload, alert.TriggeredAt)
	if err != nil {
		return fmt.Errorf("inserting alert: %w", err)
	}
	return nil
}

func (r *AlertRepo) FindByID(ctx context.Context, id string) (*model.Alert, error) {
	var alert model.Alert
	var severity string
	var payload []byte
	var publishedAt, acknowledgedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, installation_id, rule_id, severity, code, message, payload, triggered_at, published_at, acknowledged_at
		 FROM alerts WHERE id = $1`, id).
		Scan(&alert.ID, &alert.InstallationID, &alert.RuleID, &severity,
			&alert.Code, &alert.Message, &payload, &alert.TriggeredAt,
			&publishedAt, &acknowledgedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding alert: %w", err)
	}

	alert.Severity = model.AlertSeverity(severity)
	if publishedAt.Valid {
		alert.PublishedAt = &publishedAt.Time
	}
	if acknowledgedAt.Valid {
		alert.AcknowledgedAt = &acknowledgedAt.Time
	}
	if payload != nil {
		_ = json.Unmarshal(payload, &alert.Payload)
	}
	return &alert, nil
}

func (r *AlertRepo) FindByInstallation(ctx context.Context, installationID string, from, to time.Time) ([]model.Alert, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, installation_id, rule_id, severity, code, message, payload, triggered_at, published_at, acknowledged_at
		 FROM alerts
		 WHERE installation_id = $1 AND triggered_at >= $2 AND triggered_at <= $3
		 ORDER BY triggered_at DESC`,
		installationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying alerts: %w", err)
	}
	defer rows.Close()
	return r.scanAlerts(rows)
}

func (r *AlertRepo) FindUnacknowledged(ctx context.Context, installationID string) ([]model.Alert, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, installation_id, rule_id, severity, code, message, payload, triggered_at, published_at, acknowledged_at
		 FROM alerts
		 WHERE installation_id = $1 AND acknowledged_at IS NULL
		 ORDER BY triggered_at DESC`,
		installationID)
	if err != nil {
		return nil, fmt.Errorf("querying unacknowledged alerts: %w", err)
	}
	defer rows.Close()
	return r.scanAlerts(rows)
}

func (r *AlertRepo) Acknowledge(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE alerts SET acknowledged_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("acknowledging alert: %w", err)
	}
	return nil
}

func (r *AlertRepo) scanAlerts(rows *sql.Rows) ([]model.Alert, error) {
	var alerts []model.Alert
	for rows.Next() {
		var alert model.Alert
		var severity string
		var payload []byte
		var publishedAt, acknowledgedAt sql.NullTime

		err := rows.Scan(&alert.ID, &alert.InstallationID, &alert.RuleID, &severity,
			&alert.Code, &alert.Message, &payload, &alert.TriggeredAt,
			&publishedAt, &acknowledgedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning alert row: %w", err)
		}

		alert.Severity = model.AlertSeverity(severity)
		if publishedAt.Valid {
			alert.PublishedAt = &publishedAt.Time
		}
		if acknowledgedAt.Valid {
			alert.AcknowledgedAt = &acknowledgedAt.Time
		}
		if payload != nil {
			_ = json.Unmarshal(payload, &alert.Payload)
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}
