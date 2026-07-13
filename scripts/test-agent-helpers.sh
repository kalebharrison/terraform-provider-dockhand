#!/usr/bin/env bash
# shellcheck shell=bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

run() {
  echo "==> $*"
  "$@"
}

message="$(./scripts/agent-commit-msg.sh "test: helper smoke")"
run ./scripts/agent-commit-msg.sh --check <<<"${message}"

if ./scripts/agent-commit-msg.sh --check <<<"missing trailer" >/dev/null 2>&1; then
  echo "expected --check to fail without trailer" >&2
  exit 1
fi

regex="$(/usr/bin/python3 scripts/acceptance-pr-ci-regex.py)"
case "${regex}" in
  TestAcc\(*\)) ;;
  *)
    echo "unexpected PR CI regex: ${regex}" >&2
    exit 1
    ;;
esac

echo "agent helper smoke tests passed"
run /usr/bin/python3 -m unittest scripts/test_issue_agent_intake.py
run /usr/bin/python3 -m unittest scripts/test_issue_agent_intake_eligibility.py
run /usr/bin/python3 -m unittest scripts/test_issue_agent_intake_lenses.py
run /usr/bin/python3 -m unittest scripts/test_release_verdict.py
run /usr/bin/python3 -m unittest scripts/test_release_gate_check.py
run /usr/bin/python3 -m unittest scripts/test_automation_health_gate.py
run /usr/bin/python3 -m unittest scripts/test_agent_stall_watchdog.py
run /usr/bin/python3 -m unittest scripts/test_agent_approve_pr_ci.py
run /usr/bin/python3 -m unittest scripts/test_repo_hygiene.py
run /usr/bin/python3 -m unittest scripts/test_check_automation_settings.py
