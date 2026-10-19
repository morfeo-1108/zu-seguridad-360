#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

./scripts/assert-local-docker.sh
command -v curl >/dev/null || { echo 'ERROR: curl is required' >&2; exit 2; }

# Only Core's local configuration is needed for this reduced test profile.
if [[ ! -f services/core/.env.dev ]]; then
  cp services/core/.env.example services/core/.env.dev
fi
set_local_env() {
  local key="$1" value="$2" file="services/core/.env.dev"
  if grep -q "^${key}=" "$file"; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$file"
  else
    printf '\n%s=%s\n' "$key" "$value" >> "$file"
  fi
}
set_local_env ENV development
set_local_env AUTO_MIGRATE true
set_local_env PILOT_SYNTHETIC_ENABLED true
set_local_env DB_HOST postgres
set_local_env DB_NAME zu_core_synthetic

docker compose -f docker-compose.dev.yml config --quiet
echo 'Starting only the local Postgres dependency...'
docker compose -f docker-compose.dev.yml up -d postgres

ready=0
for _ in $(seq 1 60); do
  if docker compose -f docker-compose.dev.yml exec -T postgres pg_isready -U zuadmin -d zu_core_synthetic >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 2
done
if [[ "$ready" != 1 ]]; then
  docker compose -f docker-compose.dev.yml logs --tail=80 postgres
  echo 'ERROR: local Postgres did not become ready' >&2
  exit 1
fi

CONFIRM_LOCAL_TEST_DB=YES ./scripts/migrate.sh
echo 'Building and starting Core in local development mode...'
docker compose -f docker-compose.dev.yml up --build -d core

healthy=0
for _ in $(seq 1 60); do
  if curl -fsS http://127.0.0.1:4001/v1/health >/dev/null 2>&1; then
    healthy=1
    break
  fi
  sleep 2
done
if [[ "$healthy" != 1 ]]; then
  docker compose -f docker-compose.dev.yml logs --tail=120 core
  echo 'ERROR: Core health check failed; inspect logs before continuing' >&2
  exit 1
fi

./scripts/seed.sh
cat <<'EOF'
Core-only QA profile is ready with synthetic data.
Core health: http://127.0.0.1:4001/v1/health
Only Postgres and Core were requested; this does not certify other services or tenant isolation.
Use synthetic data only. Stop without deleting data: make dev-down
EOF
