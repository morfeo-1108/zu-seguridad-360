#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

cat > "$TMP/docker" <<'FAKE_DOCKER'
#!/usr/bin/env bash
set -Eeuo pipefail
case "${1:-} ${2:-}" in
  'compose version') exit 0 ;;
  'context show') printf '%s\n' "${FAKE_CONTEXT:-qa-local}" ;;
  'context inspect') printf '%s\n' "${FAKE_ENDPOINT:-unix:///var/run/docker.sock}" ;;
  *) echo "unexpected fake docker args" >&2; exit 1 ;;
esac
FAKE_DOCKER
chmod +x "$TMP/docker"

run_guard() {
  env -u DOCKER_HOST -u DOCKER_CONTEXT PATH="$TMP:$PATH" FAKE_ENDPOINT="$1" "$ROOT/scripts/assert-local-docker.sh"
}

run_guard 'unix:///var/run/docker.sock' >/dev/null
run_guard 'npipe:////./pipe/docker_engine' >/dev/null

if output="$(run_guard 'tcp://127.0.0.1:2375' 2>&1)"; then
  echo 'FAIL: TCP Docker endpoint was accepted' >&2
  exit 1
fi
grep -q 'is not local' <<<"$output" || { echo 'FAIL: TCP endpoint rejected for an unexpected reason' >&2; exit 1; }

if output="$(env -u DOCKER_CONTEXT PATH="$TMP:$PATH" DOCKER_HOST='unix:///var/run/docker.sock' FAKE_ENDPOINT='unix:///var/run/docker.sock' "$ROOT/scripts/assert-local-docker.sh" 2>&1)"; then
  echo 'FAIL: explicit DOCKER_HOST was accepted' >&2
  exit 1
fi
grep -q 'DOCKER_HOST is set' <<<"$output" || { echo 'FAIL: DOCKER_HOST rejected for an unexpected reason' >&2; exit 1; }

printf 'PASS: local Docker endpoints accepted; remote endpoints and DOCKER_HOST overrides rejected.\n'
