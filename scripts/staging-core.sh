#!/usr/bin/env bash
# Provisiona Core en ENV=staging real (commits 699138e + 99db4d0) sin relajar ValidateProduction.
# Se ejecuta por pasos: ./staging-core.sh <paso>. Cada paso verifica antes de seguir.
set -euo pipefail
cd /home/morfeo/zu-seguridad-360
P=zu-seguridad-360
NET=${P}_zu-network
DC="docker compose -f docker-compose.dev.yml"
ENVF=services/core/.env

vol() { # crea un volumen con las etiquetas de compose para que compose lo adopte sin quejarse
  docker volume inspect "${P}_$1" >/dev/null 2>&1 || docker volume create \
    --label com.docker.compose.project=$P --label com.docker.compose.volume=$1 "${P}_$1" >/dev/null
}

case "${1:-}" in
secrets) # 1. Secretos nuevos, solo en services/core/.env (gitignored), nunca se imprimen
  [ -e "$ENVF" ] && { echo "$ENVF ya existe, no lo sobrescribo"; exit 1; }
  r() { openssl rand -hex 32; }
  umask 077
  cat > "$ENVF" <<EOF
DB_PASSWORD=$(r)
REDIS_USERNAME=zu_core_staging
REDIS_PASSWORD=$(r)
RABBITMQ_URL=amqp://zu_core_staging:$(r)@rabbitmq:5672/zu_vhost
VAULT_TOKEN=$(r)
JWT_SECRET=$(r)
EOF
  # JWT_SECRET: el de .env.dev está commiteado en el repo; con usuarios reales eso permitiría falsificar tokens.
  git check-ignore -q "$ENVF" && echo "OK: $ENVF creado (600) e ignorado por git"
  ;;
certs) # 2. Certificado autofirmado para Postgres en el volumen pg_certs (uid 70 = postgres en alpine)
  vol pg_certs
  docker run --rm -v ${P}_pg_certs:/certs alpine:3.20 sh -c '
    [ -f /certs/server.key ] && { echo "certs ya existen"; exit 0; }
    apk add --no-cache openssl >/dev/null &&
    openssl req -new -x509 -days 825 -nodes -subj "/CN=zu-postgres" \
      -keyout /certs/server.key -out /certs/server.crt 2>/dev/null &&
    chown 70:70 /certs/server.key /certs/server.crt && chmod 600 /certs/server.key && ls -ln /certs'
  ;;
acl) # 3. users.acl persistente: default conserva la clave actual (billing/attention/media) + usuario de Core
  vol redis_acl
  set -a; . "$ENVF"; set +a
  docker run --rm -e CP="$REDIS_PASSWORD" -v ${P}_redis_acl:/etc/redis alpine:3.20 sh -c '
    [ -f /etc/redis/users.acl ] && { echo "users.acl ya existe"; exit 0; }
    printf "user default on >zuredis123 ~* &* +@all\nuser zu_core_staging on >%s ~* &* +@all -@admin -@dangerous\n" "$CP" > /etc/redis/users.acl &&
    chmod 644 /etc/redis/users.acl && echo "users.acl creado (2 usuarios)"'
  ;;
infra) # 4. Recrear SOLO postgres y redis con la config nueva (datos intactos en sus volúmenes)
  $DC up -d --no-deps postgres redis
  for i in $(seq 1 24); do [ "$(docker inspect -f '{{.State.Health.Status}}' zu-postgres)" = healthy ] && break; sleep 5; done
  docker exec zu-postgres psql -U zuadmin -d zu_core_synthetic -tAc "show ssl"
  docker exec zu-redis redis-cli --user default --pass zuredis123 --no-auth-warning ping
  ;;
db) # 5. Rol y base dedicados + extensiones (requieren superusuario)
  set -a; . "$ENVF"; set +a
  docker exec -i zu-postgres psql -v ON_ERROR_STOP=1 -U zuadmin -d postgres -v pw="$DB_PASSWORD" <<'SQL'
SELECT format('CREATE ROLE zu_core_staging LOGIN PASSWORD %L', :'pw')
  WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='zu_core_staging') \gexec
SELECT 'CREATE DATABASE zu_core OWNER zu_core_staging'
  WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname='zu_core') \gexec
\c zu_core
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
SQL
  ;;
rabbit) # 6. Usuario RabbitMQ propio de Core en zu_vhost
  set -a; . "$ENVF"; set +a
  RP=$(printf %s "$RABBITMQ_URL" | sed -E 's#^amqp://[^:]+:([^@]+)@.*#\1#')
  docker exec zu-rabbitmq rabbitmqctl list_users -q | grep -q '^zu_core_staging' \
    || docker exec zu-rabbitmq rabbitmqctl add_user zu_core_staging "$RP" >/dev/null
  docker exec zu-rabbitmq rabbitmqctl set_permissions -p zu_vhost zu_core_staging ".*" ".*" ".*" >/dev/null
  docker exec zu-rabbitmq rabbitmqctl list_permissions -p zu_vhost -q | grep zu_core_staging
  ;;
migrate) # 7. Migración única: GORM crea el esquema (ENV=development solo en este contenedor efímero)
  $DC build core
  docker rm -f zu-core-migrate >/dev/null 2>&1 || true
  docker run -d --name zu-core-migrate --network $NET \
    --env-file services/core/.env.dev --env-file "$ENVF" \
    -e ENV=development -e AUTO_MIGRATE=true -e PILOT_SYNTHETIC_ENABLED=false -e GOOGLE_SIGNUP_ENABLED=false \
    -e DB_HOST=postgres -e DB_USER=zu_core_staging -e DB_NAME=zu_core -e DB_SSLMODE=require \
    ${P}-core:latest >/dev/null
  for i in $(seq 1 30); do
    docker logs zu-core-migrate 2>&1 | grep -qiE "listening|started|migration failed|fatal" && break; sleep 2; done
  docker logs zu-core-migrate 2>&1 | tail -15
  docker rm -f zu-core-migrate >/dev/null
  # 001 usa CREATE INDEX CONCURRENTLY: va sin transacción; ambos idempotentes
  for f in services/core/migrations/001_initial.sql services/core/migrations/20261019_google_self_signup.sql; do
    docker exec -i zu-postgres psql -v ON_ERROR_STOP=1 -U zu_core_staging -d zu_core < "$f"; done
  docker exec zu-postgres psql -U zu_core_staging -d zu_core -tAc "select count(*) from information_schema.tables where table_schema='public'"
  ;;
core) # 8. Core en staging real
  $DC up -d --no-deps core
  sleep 15; docker ps --filter name=zu-core --format '{{.Names}} {{.Status}}'
  docker logs --tail 10 zu-core 2>&1
  curl -s http://localhost:4001/v1/auth/google/config; echo
  ;;
services) # 9. billing/attention/media: mismo JWT_SECRET que Core y la base zu_core (SSL)
  set -a; . "$ENVF"; set +a
  umask 077
  DBURL="postgres://zu_core_staging:${DB_PASSWORD}@postgres:5432/zu_core?sslmode=require"
  for s in billing attention media; do
    f=services/$s/.env
    { echo "JWT_SECRET=$JWT_SECRET"
      echo "DATABASE_URL=$DBURL"
      # "services migrate" la primera vez: billing crea sus tablas en zu_core; luego sin él.
      [ "$s" = billing ] && echo "AUTO_MIGRATE=$([ "${2:-}" = migrate ] && echo true || echo false)"
      true; } > "$f"
    git check-ignore -q "$f" && echo "OK: $f (600, ignorado por git)"
  done
  $DC up -d --no-deps billing attention media
  sleep 10; docker ps --filter name=zu-billing --filter name=zu-attention --filter name=zu-media --format '{{.Names}} {{.Status}}'
  ;;
*) echo "uso: $0 secrets|certs|acl|infra|db|rabbit|migrate|core|services"; exit 2;;
esac
