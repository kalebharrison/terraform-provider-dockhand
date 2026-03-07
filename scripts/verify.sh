#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

export GOCACHE="${GOCACHE:-$ROOT_DIR/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-$ROOT_DIR/.cache/gomod}"
mkdir -p "${GOCACHE}" "${GOMODCACHE}"

RUN_QUALITY=false
RUN_ENDPOINT_PROBE=false
RUN_ACCEPTANCE=false
TEST_REGEX="${TEST_REGEX:-TestAcc}"

usage() {
  cat <<'EOF'
Usage: ./scripts/verify.sh [options]

Options:
  --quality         Run extended quality checks (vet, golangci-lint, staticcheck, shellcheck if installed)
  --endpoint-probe  Run endpoint probe (requires Dockhand auth env vars)
  --acceptance      Run acceptance harness (Docker + Dockhand + DinD)
  --test-regex REG  Regex for acceptance tests (default: $TEST_REGEX or TestAcc)
  -h, --help        Show this help text
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --quality)
      RUN_QUALITY=true
      shift
      ;;
    --endpoint-probe)
      RUN_ENDPOINT_PROBE=true
      shift
      ;;
    --acceptance)
      RUN_ACCEPTANCE=true
      shift
      ;;
    --test-regex)
      TEST_REGEX="${2:-}"
      if [[ -z "${TEST_REGEX}" ]]; then
        echo "--test-regex requires a value" >&2
        exit 2
      fi
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
done

run() {
  echo "==> $*"
  "$@"
}

echo "==> Core checks"

tracked_go_files="$(git ls-files '*.go')"
if [[ -z "${tracked_go_files}" ]]; then
  echo "No tracked Go files found." >&2
  exit 1
fi

unformatted="$(echo "${tracked_go_files}" | xargs gofmt -l)"
if [[ -n "${unformatted}" ]]; then
  echo "gofmt check failed; unformatted files:"
  echo "${unformatted}"
  exit 1
fi

before_tidy="$(git status --porcelain -- go.mod go.sum || true)"
run go mod tidy
after_tidy="$(git status --porcelain -- go.mod go.sum || true)"
if [[ "${before_tidy}" != "${after_tidy}" ]]; then
  echo "go.mod/go.sum changed after go mod tidy. Commit tidy changes." >&2
  git --no-pager diff -- go.mod go.sum || true
  exit 1
fi

run /usr/bin/python3 scripts/check-doc-example-coverage.py
run go test ./...
run go build ./...

if [[ "${RUN_QUALITY}" == "true" ]]; then
  echo "==> Extended quality checks"
  run go vet ./...

  if command -v golangci-lint >/dev/null 2>&1; then
    run golangci-lint run
  else
    echo "warning: golangci-lint not installed; skipping."
  fi

  if ! command -v staticcheck >/dev/null 2>&1; then
    echo "==> Installing staticcheck"
    run go install honnef.co/go/tools/cmd/staticcheck@latest
    gopath="$(go env GOPATH)"
    export PATH="${PATH}:${gopath}/bin"
  fi
  run staticcheck ./...

  if command -v shellcheck >/dev/null 2>&1; then
    run shellcheck scripts/*.sh
  else
    echo "warning: shellcheck not installed; skipping."
  fi
fi

if [[ "${RUN_ENDPOINT_PROBE}" == "true" ]]; then
  echo "==> Endpoint probe"
  : "${DOCKHAND_ENDPOINT:?DOCKHAND_ENDPOINT is required for --endpoint-probe}"
  : "${DOCKHAND_USERNAME:?DOCKHAND_USERNAME is required for --endpoint-probe}"
  : "${DOCKHAND_PASSWORD:?DOCKHAND_PASSWORD is required for --endpoint-probe}"
  run /usr/bin/python3 scripts/endpoint-probe.py
fi

if [[ "${RUN_ACCEPTANCE}" == "true" ]]; then
  echo "==> Acceptance harness"
  export TEST_REGEX
  run ./scripts/run-acceptance-harness.sh
fi

echo "==> verify.sh completed successfully"
