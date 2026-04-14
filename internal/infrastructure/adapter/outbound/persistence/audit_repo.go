package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// AuditLogRepo implements port.AuditLogRepository with PostgreSQL.
type AuditLogRepo struct {
	db *sql.DB
}

func NewAuditLogRepo(db *sql.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Save(ctx context.Context, entry model.AuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (id, installation_id, action, actor, detail, timestamp)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.ID, entry.InstallationID, string(entry.Action),
		entry.Actor, entry.Detail, entry.Timestamp)
	if err != nil {
		return fmt.Errorf("inserting audit log: %w", err)
	}
	return nil
}

func (r *AuditLogRepo) FindByInstallation(ctx context.Context, installationID string, from, to time.Time) ([]model.AuditLog, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, installation_id, action, actor, detail, timestamp
		 FROM audit_logs
		 WHERE installation_id = $1 AND timestamp >= $2 AND timestamp <= $3
		 ORDER BY timestamp DESC`,
		installationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying audit logs: %w", err)
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var entry model.AuditLog
		var action string
		err := rows.Scan(&entry.ID, &entry.InstallationID, &action,
			&entry.Actor, &entry.Detail, &entry.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("scanning audit log row: %w", err)
		}
		entry.Action = model.AuditAction(action)
		logs = append(logs, entry)
	}
	return logs, rows.Err()
}
