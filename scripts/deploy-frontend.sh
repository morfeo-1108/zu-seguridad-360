#!/usr/bin/env bash
# Despliega el frontend sin bajar el sitio mientras compila:
#   1. next build en .next-build (el servidor en vivo sigue sirviendo .next)
#   2. intercambio .next → .next-prev, .next-build → .next
#   3. reinicio del proceso (systemd zu360-frontend lo relanza en ~5 s)
#   4. verificación; si falla, revierte solo.
#   ./scripts/deploy-frontend.sh            despliega
#   ./scripts/deploy-frontend.sh rollback   vuelve a la versión anterior
set -euo pipefail
cd "$(dirname "$0")/../frontend/web"
restart() {
  pid=$(pgrep -u "$(id -u)" -f "next-server" | head -1 || true)
  [ -n "$pid" ] && kill "$pid"
  for i in $(seq 1 40); do curl -sf -o /dev/null http://localhost:3000/ && return 0; sleep 1; done
  return 1
}
if [ "${1:-}" = rollback ]; then
  [ -d .next-prev ] || { echo "no hay versión anterior"; exit 1; }
  mv .next .next-failed && mv .next-prev .next && restart && echo "revertido"; exit 0
fi
rm -rf .next-build
NEXT_DIST_DIR=.next-build npx next build
rm -rf .next-prev .next-failed
mv .next .next-prev
mv .next-build .next
if restart && curl -sf -o /dev/null http://localhost:3000/dashboard; then
  echo "OK: desplegado ($(date -u +%FT%TZ)); anterior en .next-prev"
else
  echo "FALLÓ el arranque: revirtiendo"; mv .next .next-failed; mv .next-prev .next; restart; exit 1
fi
