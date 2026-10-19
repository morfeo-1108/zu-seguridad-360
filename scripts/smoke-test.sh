#!/usr/bin/env bash
set -euo pipefail
TARGET="${1:-local}"
case "$TARGET" in
  local)
    BASE_URL="${BASE_URL:-http://127.0.0.1:4001}"
    [[ "$BASE_URL" == http://127.0.0.1:* || "$BASE_URL" == http://localhost:* ]] || { echo 'Local smoke test accepts only loopback URLs' >&2; exit 2; }
    ;;
  staging)
    BASE_URL="${STAGING_CORE_URL:-}"
    [[ -n "$BASE_URL" ]] || { echo 'Set STAGING_CORE_URL in the protected staging environment' >&2; exit 2; }
    ;;
  production)
    [[ "${ALLOW_PRODUCTION_SMOKE_TEST:-}" == YES ]] || { echo 'Set ALLOW_PRODUCTION_SMOKE_TEST=YES only in an approved deployment workflow' >&2; exit 2; }
    BASE_URL="${PRODUCTION_CORE_URL:-}"
    [[ -n "$BASE_URL" ]] || { echo 'Set PRODUCTION_CORE_URL in the protected production environment' >&2; exit 2; }
    ;;
  *) echo 'Usage: smoke-test.sh local|staging|production' >&2; exit 2 ;;
esac
BASE_URL="${BASE_URL%/}"

if [[ "$TARGET" != local && "$BASE_URL" != https://* ]]; then
  echo 'Non-local smoke tests require HTTPS' >&2
  exit 2
fi

printf 'Checking verified Core health endpoint at %s/v1/health\n' "$BASE_URL"
curl --fail --silent --show-error --max-time 10 "$BASE_URL/v1/health"
printf '\nCore health endpoint passed. This smoke test does not certify other services.\n'
