#!/usr/bin/env bash
set -uo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

command -v go >/dev/null || { echo 'ERROR: Go is required (the media module declares Go 1.24.0)' >&2; exit 2; }
command -v python3 >/dev/null || { echo 'ERROR: Python 3 is required' >&2; exit 2; }

FAILED=0

echo '=== Local Docker safety guard ==='
./scripts/test-local-docker-guard.sh || FAILED=1

GO_SERVICES=(core mail notifications autopilot attention onboarding billing sla drone media)
for svc in "${GO_SERVICES[@]}"; do
  echo "=== Go tests: $svc ==="
  (cd "services/$svc" && go test ./... -count=1) || FAILED=1
done

for svc in ai documents; do
  test_dir="services/$svc/tests"
  if [[ -d "$test_dir" ]] && find "$test_dir" -type f \( -name 'test_*.py' -o -name '*_test.py' \) -print -quit | grep -q .; then
    echo "=== Python tests: $svc ==="
    (cd "services/$svc" && python3 -m pytest tests/ -v --asyncio-mode=auto) || FAILED=1
  else
    echo "=== Python tests: $svc: SKIP (no test files found) ==="
  fi
done

if [[ "$FAILED" -ne 0 ]]; then
  echo 'ONE OR MORE TEST SUITES FAILED — see output above' >&2
  exit 1
fi

echo 'ALL AVAILABLE UNIT TESTS PASSED'
