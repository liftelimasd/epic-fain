-- 001_initial_schema.sql
-- Epic-Fain: Temporal telemetry buffer + alerts + audit
-- PostgreSQL 15+

BEGIN;

-- ============================================================
-- INSTALLATIONS: Registro de instalaciones y dispositivos EPIC
-- ============================================================
CREATE TABLE IF NOT EXISTS installations (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    installation_id  VARCHAR(100) NOT NULL UNIQUE,  -- ID externo FAIN
    device_serial    VARCHAR(50),                     -- Número de serie EPCL
    epcl_type        SMALLINT DEFAULT 0,              -- Tipo de convertidor EPCL
    battery_type     SMALLINT DEFAULT 0,
    firmware_version VARCHAR(20),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_installations_serial ON installations(device_serial) WHERE device_serial IS NOT NULL;

-- ============================================================
-- TELEMETRY: Buffer temporal de datos CAN capturados
-- Reglas:
--   - Lectura NO borra
--   - Borrado solo por ACK explícito (ack_consumed)
--   - Sin UPDATE del crudo (solo delete/ack)
--   - Auditoría obligatoria en C/D
-- ============================================================
CREATE TABLE IF NOT EXISTS telemetry_records (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    installation_id  VARCHAR(100) NOT NULL REFERENCES installations(installation_id),
    message_id       INTEGER NOT NULL,                -- CAN Message ID (0xFF00, etc.)
    raw_data         BYTEA NOT NULL,                  -- Raw CAN payload
    decoded_payload  JSONB,                           -- Decoded structured data
    captured_at      TIMESTAMPTZ NOT NULL,            -- When device produced it
    received_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- When platform received it
    consumed_at      TIMESTAMPTZ                      -- NULL until ACK'd by consumer
);

-- Índices para las queries principales
CREATE INDEX idx_telemetry_installation_time
    ON telemetry_records(installation_id, received_at DESC);

CREATE INDEX idx_telemetry_unconsumed
    ON telemetry_records(installation_id, received_at)
    WHERE consumed_at IS NULL;

CREATE INDEX idx_telemetry_message_type
    ON telemetry_records(installation_id, message_id, received_at DESC);

-- Índice para purga por retención
CREATE INDEX idx_telemetry_received_at
    ON telemetry_records(received_at);

-- ============================================================
-- ALERTS: Alertas generadas por el motor de reglas
-- ============================================================
CREATE TABLE IF NOT EXISTS alerts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    installation_id  VARCHAR(100) NOT NULL REFERENCES installations(installation_id),
    rule_id          VARCHAR(100) NOT NULL,
    severity         VARCHAR(20) NOT NULL CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL')),
    code             VARCHAR(100) NOT NULL,
    message          TEXT NOT NULL,
    payload          JSONB,
    triggered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at     TIMESTAMPTZ,           -- When published to MQTT
    acknowledged_at  TIMESTAMPTZ
);

CREATE INDEX idx_alerts_installation_time
    ON alerts(installation_id, triggered_at DESC);

CREATE INDEX idx_alerts_unacknowledged
    ON alerts(installation_id, triggered_at)
    WHERE acknowledged_at IS NULL;

-- ============================================================
-- ALERT_RULES: Definición de reglas/patrones de alerta
-- Entregadas por FAIN, implementadas por Liftel
-- ============================================================
CREATE TABLE IF NOT EXISTS alert_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    message_id  INTEGER NOT NULL,
    field_name  VARCHAR(100) NOT NULL,
    operator    VARCHAR(10) NOT NULL CHECK (operator IN ('eq', 'gt', 'lt', 'gte', 'lte', 'neq', 'in')),
    value       JSONB NOT NULL,
    severity    VARCHAR(20) NOT NULL CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL')),
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- AUDIT_LOG: Trazabilidad de acciones remotas y operaciones BD
-- Quién / Cuándo / Qué / Instalación
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    installation_id  VARCHAR(100) NOT NULL,
    action           VARCHAR(50) NOT NULL,
    actor            VARCHAR(200) NOT NULL,   -- API key identifier or 'system'
    detail           TEXT,
    timestamp        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_installation_time
    ON audit_logs(installation_id, timestamp DESC);

CREATE INDEX idx_audit_action
    ON audit_logs(action, timestamp DESC);

-- ============================================================
-- CAPTURE_STATE: Estado de captura por instalación
-- ============================================================
CREATE TABLE IF NOT EXISTS capture_states (
    installation_id  VARCHAR(100) PRIMARY KEY REFERENCES installations(installation_id),
    mode             VARCHAR(20) NOT NULL DEFAULT 'NORMAL' CHECK (mode IN ('NORMAL', 'SUPERFAST')),
    active           BOOLEAN NOT NULL DEFAULT TRUE,
    superfast_end    TIMESTAMPTZ,
    measure_period_ms INTEGER NOT NULL DEFAULT 1000
);

-- ============================================================
-- API_KEYS: Autenticación por API Key (existentes en Liftel)
-- ============================================================
CREATE TABLE IF NOT EXISTS api_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash    VARCHAR(128) NOT NULL UNIQUE,  -- SHA-256 hash of the key
    name        VARCHAR(200) NOT NULL,         -- Descriptive name
    owner       VARCHAR(200) NOT NULL,         -- Who owns this key
    permissions JSONB NOT NULL DEFAULT '[]',   -- Array of allowed actions
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used   TIMESTAMPTZ
);

COMMIT;
