#!/usr/bin/env bash
# Dumps the local Compose Postgres database to a timestamped file under backups/.
# Local test/dev use only — does not target any production database.
set -Eeuo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

./scripts/assert-local-docker.sh
COMPOSE=(docker compose -f docker-compose.dev.yml)

DB_NAME="${DB_NAME:-zu_core_synthetic}"
DB_USER="${DB_USER:-zuadmin}"

"${COMPOSE[@]}" exec -T postgres pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1 || {
  echo 'ERROR: local Compose Postgres is not running' >&2
  exit 2
}

mkdir -p backups
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="backups/${DB_NAME}-${STAMP}.dump"

echo "Backing up ${DB_NAME} to ${OUT}..."
"${COMPOSE[@]}" exec -T postgres pg_dump -U "$DB_USER" -d "$DB_NAME" -Fc > "$OUT"

SIZE=$(stat -c%s "$OUT" 2>/dev/null || stat -f%z "$OUT")
if [[ "$SIZE" -lt 100 ]]; then
  echo "ERROR: backup file is suspiciously small (${SIZE} bytes); refusing to call this a success" >&2
  exit 1
fi

echo "Backup written: ${OUT} (${SIZE} bytes)"
echo "$OUT"
