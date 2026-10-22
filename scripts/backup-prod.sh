#!/usr/bin/env bash
# Backup de producción/piloto de la base zu_core (todos los esquemas: Core y
# servicios) + roles de PostgreSQL, cifrado con AES-256 (clave en
# secrets/backup.key, copia en Vault: secret/zu360/backup). Retención: 14 días.
#
#   ./scripts/backup-prod.sh            hace el backup
#   ./scripts/backup-prod.sh verify     restaura el último en una BD descartable y
#                                       compara el conteo de filas de TODAS las tablas
#
# Programado diariamente con cron (ver docs/INFRAESTRUCTURA.md). LIMITACIÓN: los
# archivos quedan en el mismo disco del host; sin copia externa, un daño del
# disco los pierde junto con la base. Copiar ZU_BACKUP_DIR fuera del host.
set -Eeuo pipefail
umask 077
cd "$(dirname "$0")/.."
DIR=${ZU_BACKUP_DIR:-$HOME/zu360-backups}
KEY=secrets/backup.key
DB=zu_core
PG() { docker exec -i zu-postgres "$@"; }
enc() { openssl enc -aes-256-cbc -pbkdf2 -iter 200000 -salt -pass "file:$KEY"; }
dec() { openssl enc -d -aes-256-cbc -pbkdf2 -iter 200000 -pass "file:$KEY"; }
log() { echo "[$(date -u +%FT%TZ)] $*"; }

ensure_key() {
  [ -f "$KEY" ] && return
  (umask 077; mkdir -p secrets; openssl rand -base64 48 > "$KEY")
  INIT=${ZU_VAULT_HOME:-$HOME/.zu360-vault}/init.json
  if [ -f "$INIT" ]; then
    VT=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "$INIT")
    docker exec -i -e VAULT_TOKEN="$VT" zu-vault vault kv put secret/zu360/backup key="$(cat "$KEY")" >/dev/null
    log "clave de cifrado nueva: $KEY (copia en Vault secret/zu360/backup)"
  fi
}

counts() { # tabla → filas, para todas las tablas de usuario de una base
  PG psql -U zuadmin -d "$1" -Atc "
    SELECT format('SELECT %L || ''|'' || count(*) FROM %I.%I;', n.nspname||'.'||c.relname, n.nspname, c.relname)
    FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE c.relkind = 'r' AND n.nspname NOT IN ('pg_catalog','information_schema') AND n.nspname NOT LIKE 'pg_toast%'
    ORDER BY 1" | PG psql -v ON_ERROR_STOP=1 -U zuadmin -d "$1" -At
}

case "${1:-backup}" in
backup)
  ensure_key
  mkdir -p "$DIR"; chmod 700 "$DIR"
  stamp=$(date -u +%Y%m%dT%H%M%SZ)
  out="$DIR/$DB-$stamp.dump.enc"
  PG pg_dump -U zuadmin -d "$DB" -Fc | enc > "$out"
  PG pg_dumpall -U zuadmin --globals-only | enc > "$DIR/globals-$stamp.sql.enc"
  size=$(stat -c%s "$out")
  [ "$size" -gt 1000 ] || { log "ERROR: backup sospechosamente pequeño ($size bytes)"; exit 1; }
  counts "$DB" > "$DIR/$DB-$stamp.counts"
  sha256sum "$out" > "$out.sha256"
  find "$DIR" -type f -mtime +14 -delete
  log "OK backup $out ($size bytes, $(wc -l < "$DIR/$DB-$stamp.counts") tablas)"
  ;;
verify)
  last=$(ls -t "$DIR"/$DB-*.dump.enc | head -1)
  sha256sum -c "$last.sha256" >/dev/null || { log "ERROR: checksum no coincide para $last"; exit 1; }
  tmp=zu360_restore_verify
  PG psql -q -U zuadmin -d postgres -c "DROP DATABASE IF EXISTS $tmp" -c "CREATE DATABASE $tmp"
  dec < "$last" | PG pg_restore -U zuadmin -d "$tmp" --no-owner --no-privileges --exit-on-error
  expected="${last%.dump.enc}.counts"
  got=$(counts "$tmp")
  if diff <(sort "$expected") <(echo "$got" | sort) >/dev/null; then
    rows=$(awk -F'|' '{s+=$2} END {print s}' "$expected")
    log "RESTORE OK: $(basename "$last") → $(wc -l < "$expected") tablas, $rows filas, conteos idénticos"
    rc=0
  else
    log "RESTORE FALLÓ: diferencias de conteo:"; diff <(sort "$expected") <(echo "$got" | sort) || true
    rc=1
  fi
  PG psql -q -U zuadmin -d postgres -c "DROP DATABASE IF EXISTS $tmp"
  exit $rc
  ;;
*) sed -n 2,13p "$0"; exit 2 ;;
esac
