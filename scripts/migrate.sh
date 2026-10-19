#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
./scripts/assert-local-docker.sh
COMPOSE=(docker compose -f docker-compose.dev.yml)


"${COMPOSE[@]}" exec -T postgres pg_isready -U zuadmin -d zu_core_synthetic >/dev/null 2>&1 || {
  echo 'ERROR: local Compose Postgres must be running' >&2
  exit 2
}

if [[ "${CONFIRM_LOCAL_TEST_DB:-}" != YES ]]; then
  if [[ ! -t 0 ]]; then
    echo 'ERROR: set CONFIRM_LOCAL_TEST_DB=YES only after verifying this is the local Compose database' >&2
    exit 2
  fi
echo 'This applies the bootstrap SQL only to local Compose database zu_core_synthetic.'
  read -r -p 'Type YES to continue: ' confirmation
  [[ "$confirmation" == YES ]] || { echo 'Cancelled.' >&2; exit 2; }
fi

"${COMPOSE[@]}" exec -T postgres \
  psql -U zuadmin -d zu_core_synthetic -v ON_ERROR_STOP=1 -f - < scripts/init-db.sql

echo 'Local test bootstrap SQL applied.'
echo 'Core model AutoMigrate runs only in development/test when AUTO_MIGRATE=true; production requires reviewed, versioned migrations.'
