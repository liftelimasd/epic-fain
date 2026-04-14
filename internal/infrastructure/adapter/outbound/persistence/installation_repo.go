package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/liftel/epic-fain/internal/domain/model"
)

// InstallationRepo implements port.InstallationRepository with PostgreSQL.
type InstallationRepo struct {
	db *sql.DB
}

func NewInstallationRepo(db *sql.DB) *InstallationRepo {
	return &InstallationRepo{db: db}
}

func (r *InstallationRepo) Save(ctx context.Context, inst model.Installation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO installations (id, installation_id, device_serial, epcl_type, battery_type, firmware_version)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		inst.ID, inst.InstallationID, inst.DeviceSerial,
		int(inst.EPCLType), int(inst.BatteryType), inst.FirmwareVer)
	if err != nil {
		return fmt.Errorf("inserting installation: %w", err)
	}
	return nil
}

func (r *InstallationRepo) FindByID(ctx context.Context, id string) (*model.Installation, error) {
	return r.scanOne(r.db.QueryRowContext(ctx,
		`SELECT id, installation_id, device_serial, epcl_type, battery_type, firmware_version, created_at, updated_at
		 FROM installations WHERE id = $1`, id))
}

func (r *InstallationRepo) FindByInstallationID(ctx context.Context, installationID string) (*model.Installation, error) {
	return r.scanOne(r.db.QueryRowContext(ctx,
		`SELECT id, installation_id, device_serial, epcl_type, battery_type, firmware_version, created_at, updated_at
		 FROM installations WHERE installation_id = $1`, installationID))
}

func (r *InstallationRepo) FindByDeviceSerial(ctx context.Context, serial string) (*model.Installation, error) {
	return r.scanOne(r.db.QueryRowContext(ctx,
		`SELECT id, installation_id, device_serial, epcl_type, battery_type, firmware_version, created_at, updated_at
		 FROM installations WHERE device_serial = $1`, serial))
}

func (r *InstallationRepo) FindAll(ctx context.Context) ([]model.Installation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, installation_id, device_serial, epcl_type, battery_type, firmware_version, created_at, updated_at
		 FROM installations ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying installations: %w", err)
	}
	defer rows.Close()

	var result []model.Installation
	for rows.Next() {
		inst, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *inst)
	}
	return result, rows.Err()
}

func (r *InstallationRepo) Update(ctx context.Context, inst model.Installation) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE installations SET device_serial = $1, epcl_type = $2, battery_type = $3,
		 firmware_version = $4, updated_at = NOW()
		 WHERE installation_id = $5`,
		inst.DeviceSerial, int(inst.EPCLType), int(inst.BatteryType),
		inst.FirmwareVer, inst.InstallationID)
	if err != nil {
		return fmt.Errorf("updating installation: %w", err)
	}
	return nil
}

func (r *InstallationRepo) scanOne(row *sql.Row) (*model.Installation, error) {
	var inst model.Installation
	var epclType, battType int
	var serial sql.NullString
	var firmware sql.NullString

	err := row.Scan(&inst.ID, &inst.InstallationID, &serial, &epclType, &battType,
		&firmware, &inst.CreatedAt, &inst.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrInstallationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning installation: %w", err)
	}

	inst.EPCLType = model.EPCLType(epclType)
	inst.BatteryType = uint8(battType)
	if serial.Valid {
		inst.DeviceSerial = serial.String
	}
	if firmware.Valid {
		inst.FirmwareVer = firmware.String
	}
	return &inst, nil
}

func (r *InstallationRepo) scanRow(rows *sql.Rows) (*model.Installation, error) {
	var inst model.Installation
	var epclType, battType int
	var serial, firmware sql.NullString

	err := rows.Scan(&inst.ID, &inst.InstallationID, &serial, &epclType, &battType,
		&firmware, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning installation row: %w", err)
	}

	inst.EPCLType = model.EPCLType(epclType)
	inst.BatteryType = uint8(battType)
	if serial.Valid {
		inst.DeviceSerial = serial.String
	}
	if firmware.Valid {
		inst.FirmwareVer = firmware.String
	}
	return &inst, nil
}
