#!/usr/bin/env bash
# Despliega los servicios del backend (docker compose) con las mismas reglas
# que el frontend:
#   1. solo desde un árbol limpio y un commit que ya esté en origin/master;
#   2. build de las imágenes y etiqueta <servicio>:<sha corto> (trazabilidad:
#      `docker image ls zu-seguridad-360-*` dice qué commit corre);
#   3. recreación de los contenedores;
#   4. verificación: contenedor en marcha, sin reinicios y /health en 200;
#      si un servicio falla, vuelve a su imagen anterior.
#
#   ./scripts/deploy-backend.sh                 despliega los 12 servicios
#   ./scripts/deploy-backend.sh billing sla     despliega solo esos
# El registro de despliegues queda en .deploy/backend.log.
set -euo pipefail
cd "$(dirname "$0")/.."

ALL=(core billing attention media mail notifications sla onboarding autopilot documents ai drone)
SVCS=("$@")
[ ${#SVCS[@]} -gt 0 ] || SVCS=("${ALL[@]}")
for s in "${SVCS[@]}"; do
  [[ " ${ALL[*]} " == *" $s "* ]] || { echo "servicio desconocido: $s"; exit 1; }
done
# drone vive en el perfil high-risk; sin el perfil compose no lo ve.
DC=(docker compose -f docker-compose.dev.yml --profile high-risk)

if [ -n "$(git status --porcelain -- services docker-compose.dev.yml infra scripts/deploy-backend.sh)" ]; then
  echo "hay cambios sin commitear en services/, docker-compose.dev.yml o infra/: commitee antes de desplegar"; exit 1
fi
rev=$(git rev-parse HEAD)
short=$(git rev-parse --short HEAD)
git fetch -q origin master
if ! git merge-base --is-ancestor "$rev" origin/master; then
  echo "el commit $rev no está en origin/master: fusione el PR (o haga push) antes de desplegar"; exit 1
fi

declare -A prev
for s in "${SVCS[@]}"; do
  prev[$s]=$(docker image inspect -f '{{.Id}}' "zu-seguridad-360-$s" 2>/dev/null || true)
done

"${DC[@]}" build "${SVCS[@]}"
for s in "${SVCS[@]}"; do docker tag "zu-seguridad-360-$s" "zu-seguridad-360-$s:$short"; done
"${DC[@]}" up -d --no-deps "${SVCS[@]}"

health() {  # $1 = servicio; 0 si /health o /v1/health responde 200
  local port
  port=$(docker port "zu-$1" 2>/dev/null | head -1 | awk -F: '{print $NF}')
  [ -n "$port" ] || return 1
  curl -sf -o /dev/null "http://127.0.0.1:$port/health" || curl -sf -o /dev/null "http://127.0.0.1:$port/v1/health"
}

failed=()
for s in "${SVCS[@]}"; do
  ok=""
  for _ in $(seq 1 60); do
    if [ "$(docker inspect -f '{{.State.Running}} {{.RestartCount}}' "zu-$s")" = "true 0" ] && health "$s"; then ok=1; break; fi
    sleep 2
  done
  if [ -n "$ok" ]; then echo "OK    $s"; else echo "FALLÓ $s"; failed+=("$s"); fi
done

mkdir -p .deploy
if [ ${#failed[@]} -gt 0 ]; then
  for s in "${failed[@]}"; do
    if [ -n "${prev[$s]}" ]; then
      docker tag "${prev[$s]}" "zu-seguridad-360-$s"
      "${DC[@]}" up -d --no-deps "$s"
      echo "revertido $s a su imagen anterior"
    fi
  done
  echo "$(date -u +%FT%TZ) $rev FALLÓ: ${failed[*]} (revertidos)" >> .deploy/backend.log
  exit 1
fi
echo "$(date -u +%FT%TZ) $rev ${SVCS[*]}" >> .deploy/backend.log
echo "OK: desplegado commit $rev en ${#SVCS[@]} servicio(s)"
