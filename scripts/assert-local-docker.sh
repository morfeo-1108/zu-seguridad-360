#!/usr/bin/env bash
set -Eeuo pipefail

fail() {
  echo "ERROR: $*" >&2
  exit 2
}

command -v docker >/dev/null 2>&1 || fail 'Docker Engine CLI is required.'
[[ -z "${DOCKER_HOST:-}" ]] || fail 'DOCKER_HOST is set; refusing to operate on an explicitly selected daemon.'

docker compose version >/dev/null 2>&1 || fail 'Docker Compose v2 plugin is required.'

context="${DOCKER_CONTEXT:-}"
if [[ -z "$context" ]]; then
  context="$(docker context show 2>/dev/null)" || fail 'Cannot determine the active Docker context.'
fi
[[ -n "$context" ]] || fail 'Docker context is empty.'
endpoint="$(docker context inspect "$context" --format '{{(index .Endpoints "docker").Host}}' 2>/dev/null)" \
  || fail "Cannot inspect Docker context '$context'."

case "$endpoint" in
  unix://*|npipe://*) ;;
  *) fail "Docker context '$context' is not local (endpoint: ${endpoint:-unknown}); only unix:// or npipe:// is allowed." ;;
esac

printf 'Docker safety check passed: context=%s, local endpoint type=%s\n' "$context" "${endpoint%%:*}"
