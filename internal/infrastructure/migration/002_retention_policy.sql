-- 002_retention_policy.sql
-- Configuración de retención parametrizable para la BD temporal.

BEGIN;

-- Tabla de configuración global del sistema
CREATE TABLE IF NOT EXISTS system_config (
    key         VARCHAR(100) PRIMARY KEY,
    value       VARCHAR(500) NOT NULL,
    description TEXT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Retención por defecto: 3 días (en horas para flexibilidad)
INSERT INTO system_config (key, value, description) VALUES
    ('telemetry.retention_hours', '72', 'Retention period for telemetry records in hours. Default: 72 (3 days)'),
    ('superfast.max_duration_seconds', '300', 'Maximum duration for superfast capture mode in seconds. Default: 300 (5 min)'),
    ('superfast.capture_period_ms', '250', 'Capture period in superfast mode. Default: 250ms'),
    ('normal.capture_period_ms', '1000', 'Capture period in normal mode. Default: 1000ms (1s)')
ON CONFLICT (key) DO NOTHING;

-- Función para purga automática (usable via pg_cron o llamada desde la app)
CREATE OR REPLACE FUNCTION purge_expired_telemetry()
RETURNS INTEGER AS $$
DECLARE
    retention_hours INTEGER;
    deleted_count INTEGER;
BEGIN
    SELECT value::INTEGER INTO retention_hours
    FROM system_config
    WHERE key = 'telemetry.retention_hours';

    IF retention_hours IS NULL THEN
        retention_hours := 72;
    END IF;

    DELETE FROM telemetry_records
    WHERE received_at < NOW() - (retention_hours || ' hours')::INTERVAL;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMIT;
