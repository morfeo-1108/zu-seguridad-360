-- ZU Seguridad 360 — Migración inicial
-- Ejecutada via GORM AutoMigrate + este seed SQL

-- Extensiones
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Índices de texto para búsqueda
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_companies_name_trgm
  ON companies USING GIN (name gin_trgm_ops);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_name_trgm
  ON employees USING GIN ((first_name || ' ' || last_name) gin_trgm_ops);

-- Índice parcial para incidentes abiertos (query más común)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_incidents_open
  ON incidents (tenant_id, company_id, occurred_at DESC)
  WHERE status = 'open' AND deleted_at IS NULL;

-- Índice para vencimientos de certificados
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_certificates_expiry
  ON certificates (expires_at, is_valid)
  WHERE is_valid = true;

-- Índice para exámenes médicos por empresa
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_medical_exams_company
  ON medical_exams (company_id, type, expires_at)
  WHERE deleted_at IS NULL;

-- Vista materializada: indicadores SST por empresa
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_company_indicators AS
SELECT
  c.id AS company_id,
  c.tenant_id,
  c.name,
  c.employee_count,
  COUNT(DISTINCT i.id) FILTER (WHERE i.type = 'accidente_trabajo' AND i.occurred_at >= NOW() - INTERVAL '12 months') AS at_count_12m,
  COUNT(DISTINCT cert.id) FILTER (WHERE cert.expires_at < NOW() AND cert.is_valid = true) AS expired_certs,
  COUNT(DISTINCT me.id) FILTER (WHERE me.expires_at < NOW()) AS expired_exams,
  c.risk_score,
  c.updated_at
FROM companies c
LEFT JOIN incidents i ON i.company_id = c.id
LEFT JOIN certificates cert ON cert.company_id = c.id
LEFT JOIN medical_exams me ON me.company_id = c.id
WHERE c.deleted_at IS NULL
GROUP BY c.id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_company_indicators ON mv_company_indicators (company_id);

-- Función para refrescar la vista
CREATE OR REPLACE FUNCTION refresh_company_indicators()
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  REFRESH MATERIALIZED VIEW CONCURRENTLY mv_company_indicators;
END;
$$;
