#!/usr/bin/env bash
# Despliega el frontend de producción sin bajar el sitio mientras compila.
#
# Producción vive en su propio clon, /srv/zu360/web, que nadie edita: el
# servicio systemd zu360-frontend corre desde /srv/zu360/web/frontend/web. Este
# script solo funciona desde ese clon y siempre despliega exactamente
# origin/master (lo fusionado en GitHub), nunca una rama ni cambios locales.
#
#   /srv/zu360/web/scripts/deploy-frontend.sh            despliega origin/master
#   /srv/zu360/web/scripts/deploy-frontend.sh rollback   vuelve a la versión anterior
#
# Pasos: 1. checkout de origin/master  2. npm ci si cambió package-lock.json
#        3. next build en .next-build (el sitio sigue sirviendo .next)
#        4. intercambio y reinicio del servicio  5. verificación; si falla, revierte.
# El commit desplegado queda en frontend/web/.next/REVISION ("<sha> <fecha UTC>").
set -euo pipefail

PROD=/srv/zu360/web
# Archivos que `next build` reescribe según el directorio de salida; no son cambios de nadie.
BUILD_TOUCHED=(frontend/web/next-env.d.ts frontend/web/tsconfig.json)

main() {
  local root
  root=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
  if [ "$root" != "$PROD" ]; then
    echo "este script solo despliega desde $PROD (se ejecutó desde $root)"; exit 1
  fi
  cd "$PROD"

  if [ "${1:-}" = rollback ]; then
    cd frontend/web
    [ -d .next-prev ] || { echo "no hay versión anterior"; exit 1; }
    mv .next .next-failed && mv .next-prev .next && restart && echo "revertido a $(cat .next/REVISION 2>/dev/null)"; exit 0
  fi

  git checkout -q -- "${BUILD_TOUCHED[@]}"
  if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
    echo "hay cambios locales en $PROD (nadie debe editar esta carpeta):"; git status --short; exit 1
  fi
  git fetch -q origin master
  git checkout -q --detach origin/master
  local rev; rev=$(git rev-parse HEAD)

  cd frontend/web
  local lock; lock=$(sha256sum package-lock.json | cut -d' ' -f1)
  if [ "$(cat node_modules/.zu-lock-sha 2>/dev/null)" != "$lock" ]; then
    npm ci --silent
    echo "$lock" > node_modules/.zu-lock-sha
  fi

  rm -rf .next-build
  NEXT_DIST_DIR=.next-build npx next build
  git -C "$PROD" checkout -q -- "${BUILD_TOUCHED[@]}"
  echo "$rev $(date -u +%FT%TZ)" > .next-build/REVISION
  rm -rf .next-prev .next-failed
  mv .next .next-prev
  mv .next-build .next
  if restart && curl -sf -o /dev/null http://localhost:3000/dashboard; then
    echo "OK: desplegado commit $rev ($(date -u +%FT%TZ)) desde $(readlink -f /proc/"$(systemctl show -p MainPID --value zu360-frontend)"/cwd); anterior en .next-prev"
  else
    echo "FALLÓ el arranque: revirtiendo"; mv .next .next-failed; mv .next-prev .next; restart; exit 1
  fi
}

# Reinicia el proceso del servicio (systemd lo relanza: Restart=always) y
# espera a que responda. Solo ese proceso: no toca servidores de desarrollo.
restart() {
  local pid
  pid=$(systemctl show -p MainPID --value zu360-frontend)
  [ "$pid" != 0 ] && kill "$pid"
  sleep 2
  for _ in $(seq 1 40); do curl -sf -o /dev/null http://localhost:3000/ && return 0; sleep 1; done
  return 1
}

main "$@"
