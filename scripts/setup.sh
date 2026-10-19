#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

./scripts/assert-local-docker.sh
command -v curl >/dev/null || { echo 'ERROR: curl is required for local health checks' >&2; exit 2; }

# Templates are local-only and are never copied over an existing user configuration.
for dir in services/*; do
  [[ -f "$dir/.env.example" ]] || continue
  [[ -f "$dir/.env.dev" ]] || cp "$dir/.env.example" "$dir/.env.dev"
done
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
echo 'Starting isolated local dependencies...'
docker compose -f docker-compose.dev.yml up -d postgres redis rabbitmq weaviate neo4j emqx minio

ready=0
for _ in $(seq 1 60); do
  if docker compose -f docker-compose.dev.yml exec -T postgres pg_isready -U zuadmin -d zu_core_synthetic >/dev/null 2>&1; then ready=1; break; fi
  sleep 2
done
if [[ "$ready" != 1 ]]; then
  docker compose -f docker-compose.dev.yml logs --tail=80 postgres
  echo 'ERROR: local Postgres did not become ready' >&2
  exit 1
fi

CONFIRM_LOCAL_TEST_DB=YES ./scripts/migrate.sh
echo 'Building and starting Core; development AutoMigrate is opt-in via the local config.'
docker compose -f docker-compose.dev.yml up --build -d core
healthy=0
for _ in $(seq 1 60); do
  if curl -fsS http://127.0.0.1:4001/v1/health >/dev/null 2>&1; then healthy=1; break; fi
  sleep 2
done
if [[ "$healthy" != 1 ]]; then
  docker compose -f docker-compose.dev.yml logs --tail=120 core
  echo 'ERROR: Core health check failed; inspect the logs before continuing' >&2
  exit 1
fi

./scripts/seed.sh
echo 'Starting remaining local backend services...'
docker compose -f docker-compose.dev.yml up --build -d
cat <<'EOF'
Local test stack started. Use synthetic data only.
Core: http://127.0.0.1:4001/v1/health
Stop without deleting test data: make dev-down
Do not expose these ports publicly or reuse local credentials outside this isolated test stack.
EOF
