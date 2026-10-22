#!/usr/bin/env bash
# Operación de Vault (modo servidor) para el despliegue de un solo host.
#
#   ./scripts/vault.sh init                 inicializa (una sola vez), dessella y monta KV v2
#   ./scripts/vault.sh unseal               dessella tras un reinicio del contenedor/host
#   ./scripts/vault.sh service <svc>        política de solo-lectura + token propio del servicio
#   ./scripts/vault.sh import <svc> <file>  carga KEY=VALUE del archivo en secret/zu360/<svc>
#   ./scripts/vault.sh keys <svc>           lista los NOMBRES de llaves guardadas (nunca valores)
#
# Material sensible (nunca en el repo):
#   $ZU_VAULT_HOME/init.json   llaves de unseal + token raíz → muévelo a un gestor de
#                              contraseñas y deja aquí solo lo necesario para dessellar
#   secrets/vault/<svc>.token  token de solo-lectura de cada servicio (montado como
#                              Docker secret en /run/secrets/vault_token)
set -euo pipefail
cd "$(dirname "$0")/.."
ZU_VAULT_HOME=${ZU_VAULT_HOME:-$HOME/.zu360-vault}
INIT=$ZU_VAULT_HOME/init.json
TOKDIR=secrets/vault
V() { docker exec -e VAULT_TOKEN="${VT:-}" zu-vault vault "$@"; }
root() { VT=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "$INIT"); }

unseal() {
  [ -f "$INIT" ] || { echo "falta $INIT (llaves de unseal)"; exit 1; }
  if { V status -format=json || true; } | python3 -c 'import json,sys;sys.exit(0 if json.load(sys.stdin)["sealed"] else 1)'; then
    for k in $(python3 -c 'import json,sys;print(" ".join(json.load(open(sys.argv[1]))["unseal_keys_b64"][:2]))' "$INIT"); do
      V operator unseal "$k" >/dev/null
    done
  fi
  V status | grep -E '^(Initialized|Sealed)'
}

case "${1:-}" in
init)
  docker compose -f docker-compose.dev.yml up -d vault >/dev/null
  # vault status: 0 = dessellado, 2 = sellado, 1 = no responde todavía
  for i in $(seq 1 20); do rc=0; V status >/dev/null 2>&1 || rc=$?; [ $rc -ne 1 ] && break; sleep 1; done
  if { V status -format=json 2>/dev/null || true; } | python3 -c 'import json,sys;sys.exit(0 if json.load(sys.stdin)["initialized"] else 1)'; then
    echo "Vault ya estaba inicializado"
  else
    (umask 077; mkdir -p "$ZU_VAULT_HOME"; V operator init -key-shares=3 -key-threshold=2 -format=json > "$INIT")
    echo "Inicializado: llaves en $INIT (600). Muévelas a un gestor de contraseñas."
  fi
  unseal
  root
  V secrets list -format=json | grep -q '"secret/"' || V secrets enable -path=secret -version=2 kv >/dev/null
  V secrets list | grep '^secret/'
  ;;
unseal) unseal ;;
service)
  svc=${2:?servicio}; root
  printf 'path "secret/data/zu360/%s" { capabilities = ["read"] }\n' "$svc" \
    | docker exec -i -e VAULT_TOKEN="$VT" zu-vault vault policy write "zu360-$svc" - >/dev/null
  (umask 077; mkdir -p "$TOKDIR"
   V token create -orphan -policy="zu360-$svc" -ttl=8760h -display-name="$svc" -field=token > "$TOKDIR/$svc.token")
  # El contenedor corre sin root en algunas imágenes: legible por el dueño y el grupo docker no.
  chmod 644 "$TOKDIR/$svc.token"
  git check-ignore -q "$TOKDIR/$svc.token" && echo "OK: política zu360-$svc + $TOKDIR/$svc.token (ignorado por git)"
  ;;
import)
  svc=${2:?servicio}; file=${3:?archivo KEY=VALUE}; root
  python3 - "$file" <<'PY' | docker exec -i -e VAULT_TOKEN="$VT" zu-vault vault kv put "secret/zu360/$svc" - >/dev/null
import json, sys
out = {}
for line in open(sys.argv[1]):
    line = line.strip()
    if not line or line.startswith("#") or "=" not in line:
        continue
    k, v = line.split("=", 1)
    out[k.strip().lower()] = v.strip()
print(json.dumps(out))
PY
  echo "OK: secret/zu360/$svc ← $(grep -cE '^[A-Za-z_]+=' "$file") llaves de $file"
  ;;
keys)
  svc=${2:?servicio}; root
  V kv get -format=json "secret/zu360/$svc" | python3 -c 'import json,sys;print(sorted(json.load(sys.stdin)["data"]["data"]))'
  ;;
*) sed -n 2,15p "$0"; exit 2 ;;
esac
