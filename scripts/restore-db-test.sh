#!/usr/bin/env bash
# Restores the most recent backup (or one passed as $1) into a throwaway database
# on the SAME local Compose Postgres instance, compares row counts against the
# source database, then drops the throwaway database. Proves a backup is
# actually restorable instead of just assuming it is.
set -Eeuo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

./scripts/assert-local-docker.sh
COMPOSE=(docker compose -f docker-compose.dev.yml)

DB_NAME="${DB_NAME:-zu_core_synthetic}"
DB_USER="${DB_USER:-zuadmin}"
RESTORE_DB="zu360_restore_test"

DUMP_FILE="${1:-}"
if [[ -z "$DUMP_FILE" ]]; then
  DUMP_FILE="$(ls -t backups/${DB_NAME}-*.dump 2>/dev/null | head -1)"
fi
[[ -n "$DUMP_FILE" && -f "$DUMP_FILE" ]] || {
  echo "ERROR: no dump file found; run scripts/backup-db.sh first or pass a path" >&2
  exit 2
}

"${COMPOSE[@]}" exec -T postgres pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1 || {
  echo 'ERROR: local Compose Postgres is not running' >&2
  exit 2
}

echo "Restoring ${DUMP_FILE} into throwaway database ${RESTORE_DB}..."
"${COMPOSE[@]}" exec -T postgres psql -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS ${RESTORE_DB};"
"${COMPOSE[@]}" exec -T postgres psql -U "$DB_USER" -d postgres -c "CREATE DATABASE ${RESTORE_DB};"
"${COMPOSE[@]}" exec -T postgres pg_restore -U "$DB_USER" -d "$RESTORE_DB" --no-owner --no-privileges < "$DUMP_FILE"

echo
echo "Comparing row counts: source (${DB_NAME}) vs restored (${RESTORE_DB})"
FAILED=0
for table in companies users company_memberships employees; do
  src=$("${COMPOSE[@]}" exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -Atc "SELECT count(*) FROM ${table}" 2>/dev/null || echo "ERR")
  dst=$("${COMPOSE[@]}" exec -T postgres psql -U "$DB_USER" -d "$RESTORE_DB" -Atc "SELECT count(*) FROM ${table}" 2>/dev/null || echo "ERR")
  if [[ "$src" == "$dst" && "$src" != "ERR" ]]; then
    echo "  PASS ${table}: ${src} rows in both"
  else
    echo "  FAIL ${table}: source=${src} restored=${dst}"
    FAILED=1
  fi
done

echo "Dropping throwaway database ${RESTORE_DB}..."
"${COMPOSE[@]}" exec -T postgres psql -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS ${RESTORE_DB};" >/dev/null

if [[ "$FAILED" -ne 0 ]]; then
  echo "RESTORE TEST FAILED" >&2
  exit 1
fi
echo "RESTORE TEST PASSED: backup is real and restorable."
