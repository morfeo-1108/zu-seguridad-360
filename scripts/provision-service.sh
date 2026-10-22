#!/usr/bin/env bash
# Aprovisiona un microservicio para staging/producción sobre Vault:
#   - usuario RabbitMQ propio con permisos limitados a sus colas/exchanges
#   - (opcional) rol PostgreSQL propio con su esquema en zu_core
#   - secretos generados aquí y escritos directo en Vault (secret/zu360/<svc>),
#     nunca en el repositorio ni en un .env
#   - token de solo-lectura del servicio (scripts/vault.sh service)
#
# Uso: ./scripts/provision-service.sh <svc> [--rotate]
# Idempotente: si el servicio ya tiene secretos en Vault no los regenera salvo --rotate.
set -euo pipefail
cd "$(dirname "$0")/.."
svc=${1:?servicio}; rotate=${2:-}
INIT=${ZU_VAULT_HOME:-$HOME/.zu360-vault}/init.json
VT=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "$INIT")
vault() { docker exec -i -e VAULT_TOKEN="$VT" zu-vault vault "$@"; }
r() { openssl rand -hex 32; }

# Permisos RabbitMQ por servicio: "configure :: write :: read" (regex sobre nombres).
# amq.default = exchange por defecto, necesario para publicar directo a una cola.
declare -A RMQ=(
  [mail]='^zu\.mail\..* :: ^(zu\.mail\..*|amq\.default)$ :: ^zu\.mail\..*'
  [notifications]='^zu\.notifications.* :: ^(zu\.notifications.*|zu\.mail\.outbound|amq\.default)$ :: ^zu\.notifications.*'
  [sla]='^zu\.sla\..* :: ^(zu\.sla\..*|zu\.notifications|amq\.default)$ :: ^(zu\.sla\..*|zu\.companies)$'
  [onboarding]='^zu\.onboarding\..* :: ^(zu\.onboarding\..*|zu\.companies|zu\.notifications|amq\.default)$ :: ^(zu\.onboarding\..*|zu\.companies)$'
  [autopilot]='^zu\.autopilot\..* :: ^(zu\.autopilot\..*|zu\.notifications|amq\.default)$ :: ^zu\.autopilot\..*'
  [documents]='^zu\.documents.* :: ^(zu\.documents.*|amq\.default)$ :: ^zu\.documents.*'
  [ai]='^zu\.ai\..* :: ^(zu\.ai\..*|amq\.default)$ :: ^zu\.ai\..*'
  [billing]='^zu\.billing.* :: ^(zu\.billing.*|zu\.notifications|amq\.default)$ :: ^zu\.billing.*'
  [attention]='^zu\.attention.* :: ^(zu\.attention.*|amq\.default)$ :: ^zu\.attention.*'
)
# Lectura (solo SELECT) de tablas de Core que cada servicio necesita. Nunca
# escritura: Core es el único dueño de esos datos.
declare -A CORE_READ=(
  [onboarding]="companies legal_requirements risk_entries epp_items committees management_systems inspections documents"
  [autopilot]="companies company_memberships inspections legal_requirements epp_items epp_deliveries committees committee_meetings management_systems sst_objectives"
)
# Tablas con columnas sensibles: solo estas columnas (users.password_hash nunca).
declare -A CORE_READ_COLS=(
  [autopilot:users]="id, tenant_id, email, first_name, role, is_active, deleted_at"
  [documents:companies]="id, tenant_id, name, nit, legal_rep, city, deleted_at"
)
# Servicios con esquema propio en zu_core (nombre de esquema = servicio).
DBSVC=" sla onboarding autopilot notifications documents ai "

existing=$(vault kv get -format=json "secret/zu360/$svc" 2>/dev/null || true)
if [ -n "$existing" ] && [ "$rotate" != "--rotate" ]; then
  echo "secret/zu360/$svc ya existe (usa --rotate para regenerar)"; exit 0
fi

tmp=$(mktemp); trap 'shred -u "$tmp" 2>/dev/null || rm -f "$tmp"' EXIT
chmod 600 "$tmp"

# ── RabbitMQ ──────────────────────────────────────────────
if [ -n "${RMQ[$svc]:-}" ]; then
  P=(); while IFS= read -r x; do P+=("$x"); done < <(echo "${RMQ[$svc]}" | sed 's/ :: /\n/g')
  rp=$(r); user="zu_$svc"
  if docker exec zu-rabbitmq rabbitmqctl list_users -q | awk '{print $1}' | grep -qx "$user"; then
    docker exec zu-rabbitmq rabbitmqctl change_password "$user" "$rp" >/dev/null
  else
    docker exec zu-rabbitmq rabbitmqctl add_user "$user" "$rp" >/dev/null
  fi
  docker exec zu-rabbitmq rabbitmqctl set_permissions -p zu_vhost "$user" "${P[0]}" "${P[1]}" "${P[2]}" >/dev/null
  echo "RABBITMQ_URL=amqp://$user:$rp@rabbitmq:5672/zu_vhost" >> "$tmp"
  echo "OK rabbitmq: $user [$(docker exec zu-rabbitmq rabbitmqctl list_user_permissions -p zu_vhost "$user" -q 2>/dev/null | head -1 | cut -f1) ...]"
fi

# ── PostgreSQL ────────────────────────────────────────────
if [[ "$DBSVC" == *" $svc "* ]]; then
  dp=$(r); role="zu_$svc"
  docker exec -i zu-postgres psql -q -v ON_ERROR_STOP=1 -U zuadmin -d zu_core -v pw="$dp" -v role="$role" -v sch="$svc" <<'SQL'
SELECT format('CREATE ROLE %I LOGIN', :'role') WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'role') \gexec
SELECT format('ALTER ROLE %I PASSWORD %L', :'role', :'pw') \gexec
SELECT format('GRANT CONNECT ON DATABASE zu_core TO %I', :'role') \gexec
SELECT format('CREATE SCHEMA IF NOT EXISTS %I AUTHORIZATION %I', :'sch', :'role') \gexec
SELECT format('ALTER ROLE %I SET search_path = %I', :'role', :'sch') \gexec
SELECT format('REVOKE CREATE ON SCHEMA public FROM %I', :'role') \gexec
SQL
  for t in ${CORE_READ[$svc]:-}; do
    docker exec zu-postgres psql -q -v ON_ERROR_STOP=1 -U zuadmin -d zu_core -c "GRANT SELECT ON public.$t TO $role" >/dev/null
  done
  for k in "${!CORE_READ_COLS[@]}"; do
    [ "${k%%:*}" = "$svc" ] || continue
    docker exec zu-postgres psql -q -v ON_ERROR_STOP=1 -U zuadmin -d zu_core -c "GRANT SELECT (${CORE_READ_COLS[$k]}) ON public.${k#*:} TO $role" >/dev/null
    echo "OK postgres: SELECT ($(echo ${CORE_READ_COLS[$k]})) sobre ${k#*:}"
  done
  [ -n "${CORE_READ[$svc]:-}" ] && echo "OK postgres: SELECT para $role sobre: ${CORE_READ[$svc]}"
  if [ "$svc" = autopilot ]; then  # el briefing incluye plazos SLA vencidos
    docker exec zu-postgres psql -q -v ON_ERROR_STOP=1 -U zuadmin -d zu_core -c "GRANT USAGE ON SCHEMA sla TO zu_autopilot; GRANT SELECT ON sla.events TO zu_autopilot" >/dev/null
  fi
  echo "DATABASE_URL=postgres://$role:$dp@postgres:5432/zu_core?sslmode=require" >> "$tmp"
  echo "OK postgres: rol $role, esquema $svc"
fi

# ── Secretos propios del servicio ─────────────────────────
case "$svc" in
  mail|notifications) echo "INTERNAL_TOKEN=$(r)" >> "$tmp" ;;
esac
# JWT_SECRET compartido con Core (valida los tokens de sesión de los usuarios).
case "$svc" in
  sla|onboarding|autopilot|notifications|documents|ai)
    jwt=$(vault kv get -field=jwt_secret secret/zu360/core 2>/dev/null || grep -E '^JWT_SECRET=' services/core/.env | cut -d= -f2-)
    echo "JWT_SECRET=$jwt" >> "$tmp" ;;
esac
# Secretos extra que el operador deja en secrets/extra/<svc>.env (p. ej. llaves de terceros).
[ -f "secrets/extra/$svc.env" ] && grep -E '^[A-Z_]+=.' "secrets/extra/$svc.env" >> "$tmp"

./scripts/vault.sh import "$svc" "$tmp"
./scripts/vault.sh service "$svc"
./scripts/vault.sh keys "$svc"
