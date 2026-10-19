#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLEAN_CACHE=false

usage() {
  cat <<'EOF'
Uso: scripts/sync-go-deps.sh [--clean-cache]

  --clean-cache  Ejecuta `go clean -modcache` una sola vez antes de procesar
                 los módulos. Borra la caché global de módulos Go del usuario;
                 las dependencias deberán descargarse otra vez.
EOF
}

for arg in "$@"; do
  case "$arg" in
    --clean-cache) CLEAN_CACHE=true ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Argumento desconocido: $arg" >&2; usage >&2; exit 2 ;;
  esac
done

command -v go >/dev/null || { echo 'ERROR: Go no está instalado' >&2; exit 2; }
mapfile -t MODFILES < <(find "$ROOT/services" -mindepth 2 -maxdepth 2 -type f -name go.mod | sort)
((${#MODFILES[@]} > 0)) || { echo 'ERROR: no se encontraron services/*/go.mod' >&2; exit 2; }

if [[ "$CLEAN_CACHE" == true ]]; then
  echo 'Limpiando caché global de módulos Go...'
  go clean -modcache
fi

printf 'Módulos Go encontrados: %d\n' "${#MODFILES[@]}"
for modfile in "${MODFILES[@]}"; do
  module_dir="$(dirname "$modfile")"
  service="${module_dir##*/}"
  echo "=== $service: go mod tidy ==="
  (
    cd "$module_dir"
    GOTOOLCHAIN=auto go mod tidy
    test -s go.sum || { echo "ERROR: $service no generó go.sum" >&2; exit 1; }
    GOTOOLCHAIN=auto go mod verify
    echo "=== $service: go build ./... ==="
    GOTOOLCHAIN=auto go build ./...
  )
done

echo 'PASS: todos los módulos Go resolvieron dependencias, verificaron módulos y compilaron.'
