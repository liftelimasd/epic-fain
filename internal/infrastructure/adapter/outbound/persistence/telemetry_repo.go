package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// TelemetryRepo implements port.TelemetryRepository with PostgreSQL.
type TelemetryRepo struct {
	db *sql.DB
}

func NewTelemetryRepo(db *sql.DB) *TelemetryRepo {
	return &TelemetryRepo{db: db}
}

func (r *TelemetryRepo) Store(ctx context.Context, record model.TelemetryRecord) error {
	payload, err := json.Marshal(record.DecodedPayload)
	if err != nil {
		return fmt.Errorf("marshaling decoded payload: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO telemetry_records (id, installation_id, message_id, raw_data, decoded_payload, captured_at, received_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		record.ID,
		record.InstallationID,
		int(record.MessageID),
		record.RawData,
		payload,
		record.CapturedAt,
		record.ReceivedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting telemetry record: %w", err)
	}
	return nil
}

func (r *TelemetryRepo) FindByID(ctx context.Context, id string) (*model.TelemetryRecord, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, installation_id, message_id, raw_data, decoded_payload, captured_at, received_at, consumed_at
		 FROM telemetry_records WHERE id = $1`, id)
	return r.scanRecord(row)
}

func (r *TelemetryRepo) FindByInstallation(ctx context.Context, installationID string, from, to time.Time) ([]model.TelemetryRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, installation_id, message_id, raw_data, decoded_payload, captured_at, received_at, consumed_at
		 FROM telemetry_records
		 WHERE installation_id = $1 AND received_at >= $2 AND received_at <= $3
		 ORDER BY received_at DESC`,
		installationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying telemetry by installation: %w", err)
	}
	defer rows.Close()
	return r.scanRecords(rows)
}

func (r *TelemetryRepo) FindUnconsumed(ctx context.Context, installationID string, limit int) ([]model.TelemetryRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, installation_id, message_id, raw_data, decoded_payload, captured_at, received_at, consumed_at
		 FROM telemetry_records
		 WHERE installation_id = $1 AND consumed_at IS NULL
		 ORDER BY received_at ASC
		 LIMIT $2`,
		installationID, limit)
	if err != nil {
		return nil, fmt.Errorf("querying unconsumed telemetry: %w", err)
	}
	defer rows.Close()
	return r.scanRecords(rows)
}

func (r *TelemetryRepo) AckConsumed(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE telemetry_records SET consumed_at = NOW() WHERE id = $1 AND consumed_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("acking telemetry record: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return model.ErrAlreadyConsumed
	}
	return nil
}

func (r *TelemetryRepo) DeleteConsumed(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	// Build parameterized query for variable number of IDs
	query := `DELETE FROM telemetry_records WHERE id = ANY($1) AND consumed_at IS NOT NULL`
	_, err := r.db.ExecContext(ctx, query, pqStringArray(ids))
	if err != nil {
		return fmt.Errorf("deleting consumed telemetry records: %w", err)
	}
	return nil
}

func (r *TelemetryRepo) PurgeExpired(ctx context.Context, retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM telemetry_records WHERE received_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purging expired telemetry: %w", err)
	}
	return result.RowsAffected()
}

// scanRecord scans a single row into a TelemetryRecord.
func (r *TelemetryRepo) scanRecord(row *sql.Row) (*model.TelemetryRecord, error) {
	var rec model.TelemetryRecord
	var msgID int
	var payload []byte
	var consumedAt sql.NullTime

	err := row.Scan(&rec.ID, &rec.InstallationID, &msgID, &rec.RawData, &payload,
		&rec.CapturedAt, &rec.ReceivedAt, &consumedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrTelemetryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning telemetry record: %w", err)
	}

	rec.MessageID = model.MessageID(msgID)
	if consumedAt.Valid {
		rec.ConsumedAt = &consumedAt.Time
	}
	// DecodedPayload left as raw JSON; callers decode as needed
	return &rec, nil
}

// scanRecords scans multiple rows into a slice of TelemetryRecord.
func (r *TelemetryRepo) scanRecords(rows *sql.Rows) ([]model.TelemetryRecord, error) {
	var records []model.TelemetryRecord
	for rows.Next() {
		var rec model.TelemetryRecord
		var msgID int
		var payload []byte
		var consumedAt sql.NullTime

		err := rows.Scan(&rec.ID, &rec.InstallationID, &msgID, &rec.RawData, &payload,
			&rec.CapturedAt, &rec.ReceivedAt, &consumedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning telemetry rows: %w", err)
		}

		rec.MessageID = model.MessageID(msgID)
		if consumedAt.Valid {
			rec.ConsumedAt = &consumedAt.Time
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// pqStringArray converts a []string to a PostgreSQL text array literal.
func pqStringArray(ss []string) string {
	result := "{"
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += `"` + s + `"`
	}
	result += "}"
	return result
}
