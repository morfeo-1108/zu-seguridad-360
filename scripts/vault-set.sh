#!/usr/bin/env bash
# Guarda uno o más secretos de un servicio en Vault sin que pasen por archivos,
# historial de shell ni salida en pantalla, y reinicia el servicio.
#
#   ./scripts/vault-set.sh billing wompi_public_key wompi_private_key wompi_events_secret wompi_integrity_secret
#
# Pide cada valor con entrada oculta. Las llaves pub_test_/prv_test_ activan el
# sandbox de Wompi; pub_prod_/prv_prod_ producción (no hace falta cambiar código).
set -euo pipefail
cd "$(dirname "$0")/.."
svc=${1:?servicio}; shift
[ $# -gt 0 ] || { echo "indique al menos una llave"; exit 2; }
VT=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "${ZU_VAULT_HOME:-$HOME/.zu360-vault}/init.json")
args=()
for k in "$@"; do
  read -rsp "$k: " v; echo
  [ -n "$v" ] || { echo "valor vacío para $k"; exit 1; }
  args+=("$k=$v")
done
printf '%s\n' "${args[@]}" | python3 -c 'import json,sys; print(json.dumps(dict(l.rstrip("\n").split("=",1) for l in sys.stdin if l.strip())))' \
  | docker exec -i -e VAULT_TOKEN="$VT" zu-vault vault kv patch "secret/zu360/$svc" - >/dev/null
echo "OK: $# llave(s) guardadas en secret/zu360/$svc"
docker compose -f docker-compose.dev.yml up -d --no-deps --force-recreate "$svc" 2>&1 | tail -1
