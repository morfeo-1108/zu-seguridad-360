#!/usr/bin/env bash
# Rota el JWT_SECRET compartido (firma de las sesiones de Core, validado por los
# demás servicios). Todas las sesiones activas quedan inválidas: los usuarios
# vuelven a iniciar sesión. Uso: ./scripts/rotate-jwt-secret.sh
set -euo pipefail
cd "$(dirname "$0")/.."
VT=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "${ZU_VAULT_HOME:-$HOME/.zu360-vault}/init.json")
V() { docker exec -i -e VAULT_TOKEN="$VT" zu-vault vault "$@"; }
new=$(openssl rand -hex 32)
svcs=()
for s in $(V kv list -format=json secret/zu360 | python3 -c 'import json,sys;print(" ".join(json.load(sys.stdin)))'); do
  if V kv get -field=jwt_secret "secret/zu360/$s" >/dev/null 2>&1; then
    V kv patch "secret/zu360/$s" jwt_secret="$new" >/dev/null
    svcs+=("$s")
  fi
done
echo "jwt_secret rotado en: ${svcs[*]}"
running=()
for s in "${svcs[@]}"; do docker ps --format '{{.Names}}' | grep -qx "zu-$s" && running+=("$s"); done
docker compose -f docker-compose.dev.yml up -d --no-deps --force-recreate "${running[@]}" 2>&1 | grep -c Started | xargs echo "reiniciados:"
