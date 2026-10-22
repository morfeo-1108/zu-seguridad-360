#!/usr/bin/env bash
# Tenant de QA aislado para pruebas de punta a punta contra el entorno en vivo.
# Datos 100% ficticios y marcados (metadata.qa=true, nombre "ZU QA — …"); ningún
# usuario del piloto puede verlo porque vive en su propio tenant_id.
#
#   ./scripts/qa-tenant.sh create   crea tenant, profesional SST y empresa ficticia
#   ./scripts/qa-tenant.sh login    imprime un access token real emitido por Core
#   ./scripts/qa-tenant.sh delete   borra todo lo del tenant QA (Core + esquemas de servicios)
#
# La contraseña del usuario QA se genera aquí y se guarda solo en Vault (secret/zu360/qa).
set -euo pipefail
cd "$(dirname "$0")/.."
TENANT=0a0a0a0a-0000-4000-8000-000000000001
USER_ID=0a0a0a0a-0000-4000-8000-0000000000a1
COMPANY_ID=0a0a0a0a-0000-4000-8000-0000000000c1
EMAIL=qa.sst@zuseguridad360.com
INIT=${ZU_VAULT_HOME:-$HOME/.zu360-vault}/init.json
VT=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "$INIT")
vault() { docker exec -i -e VAULT_TOKEN="$VT" zu-vault vault "$@"; }
psql_admin() { docker exec -i zu-postgres psql -q -v ON_ERROR_STOP=1 -U zuadmin -d zu_core "$@"; }

case "${1:-}" in
create)
  pw=$(vault kv get -field=password secret/zu360/qa 2>/dev/null || true)
  if [ -z "$pw" ]; then pw=$(openssl rand -base64 24 | tr -d '/+='); vault kv put secret/zu360/qa password="$pw" email="$EMAIL" >/dev/null; fi
  psql_admin -v t="$TENANT" -v u="$USER_ID" -v c="$COMPANY_ID" -v e="$EMAIL" -v pw="$pw" <<'SQL'
INSERT INTO companies (id, tenant_id, name, nit, ciiu, risk_class, arl, city, department, employee_count, status, metadata, created_at, updated_at)
VALUES (:'c', :'t', 'ZU QA — Constructora Ficticia S.A.S.', '900000001-0', '4111', 'V', 'ARL de prueba (ficticia)', 'Bogotá', 'Cundinamarca', 12, 'active', '{"qa": true, "ficticia": true}', now(), now())
ON CONFLICT (id) DO NOTHING;
INSERT INTO users (id, tenant_id, company_id, email, password_hash, auth_provider, role, first_name, last_name, is_active, terms_accepted_at, privacy_acknowledged_at, terms_version, privacy_notice_version, preferences, created_at, updated_at)
VALUES (:'u', :'t', :'c', :'e', crypt(:'pw', gen_salt('bf', 12)), 'password', 'professional_sst', 'QA', 'Profesional SST (prueba)', true, now(), now(), '1.0-piloto', '1.0-piloto', '{"qa": true}', now(), now())
ON CONFLICT (id) DO UPDATE SET password_hash = EXCLUDED.password_hash;
INSERT INTO company_memberships (tenant_id, user_id, company_id, role, is_active, created_at, updated_at)
SELECT :'t', :'u', :'c', 'manager', true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM company_memberships WHERE user_id = :'u' AND company_id = :'c');
SQL
  echo "OK: tenant QA $TENANT (usuario $EMAIL, empresa ficticia $COMPANY_ID)"
  ;;
login)
  pw=$(vault kv get -field=password secret/zu360/qa)
  python3 - "$EMAIL" "$pw" <<'PY'
import json, sys, urllib.request
req = urllib.request.Request("http://localhost:4001/v1/auth/login", method="POST",
    data=json.dumps({"email": sys.argv[1], "password": sys.argv[2]}).encode(), headers={"Content-Type": "application/json"})
d = json.load(urllib.request.urlopen(req))
print(d.get("access_token") or d["data"]["access_token"])
PY
  ;;
delete)
  for s in sla onboarding autopilot notifications; do
    psql_admin -c "DO \$\$ BEGIN IF to_regclass('$s.events') IS NOT NULL AND '$s'='sla' THEN DELETE FROM sla.events WHERE tenant_id='$TENANT'; END IF; END \$\$;" 2>/dev/null || true
  done
  psql_admin -v t="$TENANT" <<'SQL'
DELETE FROM inspections WHERE tenant_id = :'t';
DELETE FROM company_memberships WHERE tenant_id = :'t';
DELETE FROM users WHERE tenant_id = :'t';
DELETE FROM companies WHERE tenant_id = :'t';
SQL
  echo "OK: tenant QA eliminado de Core"
  ;;
*) sed -n 2,10p "$0"; exit 2 ;;
esac
