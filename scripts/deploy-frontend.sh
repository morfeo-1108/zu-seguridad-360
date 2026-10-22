#!/usr/bin/env bash
# Despliega el frontend sin bajar el sitio mientras compila:
#   1. next build en .next-build (el servidor en vivo sigue sirviendo .next)
#   2. intercambio .next → .next-prev, .next-build → .next
#   3. reinicio del proceso (systemd zu360-frontend lo relanza en ~5 s)
#   4. verificación; si falla, revierte solo.
# Solo despliega un commit limpio que ya esté en origin/master.
# El commit desplegado queda en frontend/web/.next/REVISION ("<sha> <fecha UTC>").
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
# Solo se despliega un commit: con cambios sin commitear no habría forma de
# saber después qué código está sirviendo el dominio.
if [ -n "$(git status --porcelain -- . ../../scripts/deploy-frontend.sh)" ]; then
  echo "hay cambios sin commitear en frontend/web: commitee antes de desplegar"; exit 1
fi
rev=$(git rev-parse HEAD)
# Y solo un commit que ya esté en GitHub (origin/master): el código que sirve
# el dominio no puede existir únicamente en esta máquina.
git fetch -q origin master
if ! git merge-base --is-ancestor "$rev" origin/master; then
  echo "el commit $rev no está en origin/master: fusione el PR (o haga push) antes de desplegar"; exit 1
fi
rm -rf .next-build
NEXT_DIST_DIR=.next-build npx next build
echo "$rev $(date -u +%FT%TZ)" > .next-build/REVISION
rm -rf .next-prev .next-failed
mv .next .next-prev
mv .next-build .next
if restart && curl -sf -o /dev/null http://localhost:3000/dashboard; then
  echo "OK: desplegado commit $rev ($(date -u +%FT%TZ)); anterior en .next-prev"
else
  echo "FALLÓ el arranque: revirtiendo"; mv .next .next-failed; mv .next-prev .next; restart; exit 1
fi
