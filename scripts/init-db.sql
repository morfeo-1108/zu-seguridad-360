-- Bootstrap SQL for the local development/test database only.
-- Core domain tables are created by GORM AutoMigrate in development when explicitly enabled.
-- This file intentionally creates NO users, tenants, companies, medical data, or production records.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Marker required by the synthetic seed script; this is not a production schema.
CREATE TABLE IF NOT EXISTS public.zu360_local_test_marker (
  singleton   BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton IS TRUE),
  marker_value TEXT NOT NULL CHECK (marker_value = 'zu360-local-test-only'),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO public.zu360_local_test_marker (singleton, marker_value)
VALUES (TRUE, 'zu360-local-test-only')
ON CONFLICT (singleton) DO NOTHING;

CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.event_log (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL,
  user_id     UUID,
  action      VARCHAR(100) NOT NULL,
  resource    VARCHAR(100),
  resource_id UUID,
  payload     JSONB,
  ip_address  VARCHAR(45),
  user_agent  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Defense in depth for local pilot tests. Production must apply an equivalent
-- reviewed migration and run the app as a non-owner, non-superuser role.
CREATE OR REPLACE FUNCTION audit.reject_event_log_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  RAISE EXCEPTION 'audit.event_log is append-only' USING ERRCODE = '55000';
END;
$$;

DROP TRIGGER IF EXISTS audit_event_log_no_update_delete ON audit.event_log;
CREATE TRIGGER audit_event_log_no_update_delete
  BEFORE UPDATE OR DELETE ON audit.event_log
  FOR EACH ROW EXECUTE FUNCTION audit.reject_event_log_mutation();

DROP TRIGGER IF EXISTS audit_event_log_no_truncate ON audit.event_log;
CREATE TRIGGER audit_event_log_no_truncate
  BEFORE TRUNCATE ON audit.event_log
  FOR EACH STATEMENT EXECUTE FUNCTION audit.reject_event_log_mutation();

REVOKE UPDATE, DELETE, TRUNCATE ON audit.event_log FROM PUBLIC;
CREATE INDEX IF NOT EXISTS idx_audit_tenant ON audit.event_log(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit.event_log(user_id, created_at DESC);

-- Casbin adapter table for local development/test only.
CREATE TABLE IF NOT EXISTS casbin_rule (
  id BIGSERIAL PRIMARY KEY,
  ptype VARCHAR(100),
  v0 VARCHAR(100), v1 VARCHAR(100), v2 VARCHAR(100),
  v3 VARCHAR(100), v4 VARCHAR(100), v5 VARCHAR(100)
);
CREATE INDEX IF NOT EXISTS idx_casbin ON casbin_rule(ptype, v0, v1, v2);
