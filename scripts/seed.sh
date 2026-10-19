#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
./scripts/assert-local-docker.sh
COMPOSE=(docker compose -f docker-compose.dev.yml)

"${COMPOSE[@]}" exec -T postgres pg_isready -U zuadmin -d zu_core_synthetic >/dev/null 2>&1 || {
  echo 'ERROR: only the local Compose Postgres may be seeded, and it must be running' >&2
  exit 2
}

marker=$("${COMPOSE[@]}" exec -T postgres psql -U zuadmin -d zu_core_synthetic -Atc \
  "SELECT marker_value FROM zu360_local_test_marker WHERE singleton = TRUE" 2>/dev/null || true)
if [[ "$marker" != 'zu360-local-test-only' ]]; then
  echo 'ERROR: local-test database marker is missing; run the local migration first' >&2
  exit 2
fi

if [[ "${CONFIRM_LOCAL_SYNTHETIC_SEED:-}" != YES ]]; then
  if [[ ! -t 0 ]]; then
    echo 'ERROR: set CONFIRM_LOCAL_SYNTHETIC_SEED=YES only after verifying the local Compose database' >&2
    exit 2
  fi
  echo 'This writes ONLY two synthetic SST professionals, four synthetic companies, four memberships and four synthetic worker profiles to local Compose Postgres.'
  read -r -p 'Type YES to continue: ' confirmation
  [[ "$confirmation" == YES ]] || { echo 'Cancelled.' >&2; exit 2; }
fi

TEST_SST_A_EMAIL="${TEST_SST_A_EMAIL:-sst-a@zu360.test}"
TEST_SST_B_EMAIL="${TEST_SST_B_EMAIL:-sst-b@zu360.test}"
for email in "$TEST_SST_A_EMAIL" "$TEST_SST_B_EMAIL"; do
  [[ "$email" =~ ^[A-Za-z0-9._%+-]+@zu360\.test$ ]] || {
    echo 'ERROR: test account emails must use the reserved @zu360.test domain' >&2; exit 2;
  }
done
[[ "$TEST_SST_A_EMAIL" != "$TEST_SST_B_EMAIL" ]] || { echo 'ERROR: the two test emails must differ' >&2; exit 2; }

read_password() {
  local variable_name="$1" prompt="$2" value="${!1:-}"
  if [[ -z "$value" ]]; then
    if [[ ! -t 0 ]]; then
      echo "Set $variable_name in the local environment or run interactively; do not send it in chat." >&2
      exit 2
    fi
    read -r -s -p "$prompt (min 16 characters): " value
    echo
  fi
  if (( ${#value} < 16 )); then
    echo "ERROR: $variable_name must contain at least 16 characters" >&2
    exit 2
  fi
  printf '%s' "$value"
}
TEST_SST_A_PASSWORD="$(read_password TEST_SST_A_PASSWORD 'Password for synthetic professional A')"
TEST_SST_B_PASSWORD="$(read_password TEST_SST_B_PASSWORD 'Password for synthetic professional B')"
[[ "$TEST_SST_A_PASSWORD" != "$TEST_SST_B_PASSWORD" ]] || { echo 'ERROR: use two different test passwords' >&2; exit 2; }

# Passwords are base64-encoded and passed to psql through stdin, never as CLI args.
PW_A_B64="$(printf '%s' "$TEST_SST_A_PASSWORD" | base64 | tr -d '\n')"
PW_B_B64="$(printf '%s' "$TEST_SST_B_PASSWORD" | base64 | tr -d '\n')"
{
  printf "\\set email_a '%s'\n" "$TEST_SST_A_EMAIL"
  printf "\\set email_b '%s'\n" "$TEST_SST_B_EMAIL"
  printf "\\set password_a_b64 '%s'\n" "$PW_A_B64"
  printf "\\set password_b_b64 '%s'\n" "$PW_B_B64"
  cat <<'SQL'
BEGIN;

INSERT INTO companies (id, tenant_id, name, nit, risk_class, status, is_synthetic, created_at, updated_at)
VALUES
  ('22222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111', 'ZU360 PILOTO SINTÉTICO - EMPRESA A1', 'TEST-NIT-0001', 'TEST', 'active', TRUE, NOW(), NOW()),
  ('66666666-6666-4666-8666-666666666666', '11111111-1111-4111-8111-111111111111', 'ZU360 PILOTO SINTÉTICO - EMPRESA A2', 'TEST-NIT-0003', 'TEST', 'active', TRUE, NOW(), NOW()),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'ZU360 PILOTO SINTÉTICO - EMPRESA B1', 'TEST-NIT-0002', 'TEST', 'active', TRUE, NOW(), NOW()),
  ('cccccccc-cccc-4ccc-8ccc-cccccccccccc', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'ZU360 PILOTO SINTÉTICO - EMPRESA B2', 'TEST-NIT-0004', 'TEST', 'active', TRUE, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  nit = EXCLUDED.nit,
  risk_class = EXCLUDED.risk_class,
  status = EXCLUDED.status,
  is_synthetic = TRUE,
  updated_at = NOW()
WHERE companies.tenant_id = EXCLUDED.tenant_id;

INSERT INTO users (id, tenant_id, company_id, email, password_hash, role, first_name, last_name, is_active, is_synthetic, created_at, updated_at)
VALUES
  ('33333333-3333-4333-8333-333333333333', '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222', :'email_a', crypt(convert_from(decode(:'password_a_b64', 'base64'), 'UTF8'), gen_salt('bf')), 'professional_sst', 'Profesional', 'Sintético A', TRUE, TRUE, NOW(), NOW()),
  ('44444444-4444-4444-8444-444444444444', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', :'email_b', crypt(convert_from(decode(:'password_b_b64', 'base64'), 'UTF8'), gen_salt('bf')), 'professional_sst', 'Profesional', 'Sintético B', TRUE, TRUE, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  tenant_id = EXCLUDED.tenant_id,
  company_id = EXCLUDED.company_id,
  email = EXCLUDED.email,
  password_hash = EXCLUDED.password_hash,
  role = EXCLUDED.role,
  first_name = EXCLUDED.first_name,
  last_name = EXCLUDED.last_name,
  is_active = TRUE,
  is_synthetic = TRUE,
  updated_at = NOW()
WHERE users.tenant_id = EXCLUDED.tenant_id;

INSERT INTO company_memberships (tenant_id, user_id, company_id, role, is_active, is_synthetic, created_at, updated_at)
VALUES
  ('11111111-1111-4111-8111-111111111111', '33333333-3333-4333-8333-333333333333', '22222222-2222-4222-8222-222222222222', 'manager', TRUE, TRUE, NOW(), NOW()),
  ('11111111-1111-4111-8111-111111111111', '33333333-3333-4333-8333-333333333333', '66666666-6666-4666-8666-666666666666', 'manager', TRUE, TRUE, NOW(), NOW()),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '44444444-4444-4444-8444-444444444444', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 'manager', TRUE, TRUE, NOW(), NOW()),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '44444444-4444-4444-8444-444444444444', 'cccccccc-cccc-4ccc-8ccc-cccccccccccc', 'manager', TRUE, TRUE, NOW(), NOW())
ON CONFLICT (tenant_id, user_id, company_id) DO UPDATE SET
  role = EXCLUDED.role,
  is_active = TRUE,
  is_synthetic = TRUE,
  updated_at = NOW();

INSERT INTO employees (id, tenant_id, company_id, first_name, last_name, document_type, document_number, email, position, department, contract_type, status, is_synthetic, created_at, updated_at)
VALUES
  ('a1111111-1111-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222', 'Persona', 'Sintética A1', 'SYN', 'SYN-A1-0001', 'persona.a1.01@zu360.test', 'Demostración SST', 'Sintético', 'service', 'active', TRUE, NOW(), NOW()),
  ('a2222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111', '66666666-6666-4666-8666-666666666666', 'Persona', 'Sintética A2', 'SYN', 'SYN-A2-0001', 'persona.a2.01@zu360.test', 'Demostración SST', 'Sintético', 'service', 'active', TRUE, NOW(), NOW()),
  ('b1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 'Persona', 'Sintética B1', 'SYN', 'SYN-B1-0001', 'persona.b1.01@zu360.test', 'Demostración SST', 'Sintético', 'service', 'active', TRUE, NOW(), NOW()),
  ('b2222222-2222-4222-8222-222222222222', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'cccccccc-cccc-4ccc-8ccc-cccccccccccc', 'Persona', 'Sintética B2', 'SYN', 'SYN-B2-0001', 'persona.b2.01@zu360.test', 'Demostración SST', 'Sintético', 'service', 'active', TRUE, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  tenant_id = EXCLUDED.tenant_id,
  company_id = EXCLUDED.company_id,
  first_name = EXCLUDED.first_name,
  last_name = EXCLUDED.last_name,
  document_type = EXCLUDED.document_type,
  document_number = EXCLUDED.document_number,
  email = EXCLUDED.email,
  position = EXCLUDED.position,
  department = EXCLUDED.department,
  contract_type = EXCLUDED.contract_type,
  status = EXCLUDED.status,
  is_synthetic = TRUE,
  updated_at = NOW()
WHERE employees.tenant_id = EXCLUDED.tenant_id AND employees.company_id = EXCLUDED.company_id;

COMMIT;
SQL
} | "${COMPOSE[@]}" exec -T postgres psql -U zuadmin -d zu_core_synthetic -q -v ON_ERROR_STOP=1

verified_users=$("${COMPOSE[@]}" exec -T postgres psql -U zuadmin -d zu_core_synthetic -Atc \
  "SELECT count(*) FROM users WHERE is_synthetic = TRUE AND (id, tenant_id, company_id, role) IN (('33333333-3333-4333-8333-333333333333'::uuid,'11111111-1111-4111-8111-111111111111'::uuid,'22222222-2222-4222-8222-222222222222'::uuid,'professional_sst'),('44444444-4444-4444-8444-444444444444'::uuid,'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'::uuid,'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb'::uuid,'professional_sst'))")
verified_companies=$("${COMPOSE[@]}" exec -T postgres psql -U zuadmin -d zu_core_synthetic -Atc \
  "SELECT count(*) FROM company_memberships m JOIN companies c ON c.id=m.company_id AND c.tenant_id=m.tenant_id WHERE m.is_active=TRUE AND m.is_synthetic=TRUE AND c.is_synthetic=TRUE AND (m.user_id,m.tenant_id) IN (('33333333-3333-4333-8333-333333333333'::uuid,'11111111-1111-4111-8111-111111111111'::uuid),('44444444-4444-4444-8444-444444444444'::uuid,'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'::uuid))")
verified_employees=$("${COMPOSE[@]}" exec -T postgres psql -U zuadmin -d zu_core_synthetic -Atc \
  "SELECT count(*) FROM employees WHERE is_synthetic=TRUE AND document_number LIKE 'SYN-%'")
if [[ "$verified_users" != 2 || "$verified_companies" != 4 || "$verified_employees" != 4 ]]; then
  echo 'ERROR: seed verification failed; inspect local test database before retrying' >&2
  exit 1
fi

unset TEST_SST_A_PASSWORD TEST_SST_B_PASSWORD PW_A_B64 PW_B_B64
printf 'Seed verified: two synthetic professionals, four companies, four memberships and four synthetic worker profiles across isolated tenants (%s, %s).\n' "$TEST_SST_A_EMAIL" "$TEST_SST_B_EMAIL"
