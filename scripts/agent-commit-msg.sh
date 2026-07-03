#!/usr/bin/env bash
# shellcheck shell=bash
# Print or build a commit message with the standard agent co-author trailer.
set -euo pipefail

AGENT_CO_AUTHOR_TRAILER="${AGENT_CO_AUTHOR_TRAILER:-Co-authored-by: Cursor Agent <noreply@cursor.com>}"

usage() {
  cat <<'EOF'
Usage:
  ./scripts/agent-commit-msg.sh "commit subject"
  ./scripts/agent-commit-msg.sh --check <<'MSG'
  subject line

  optional body
  MSG

Prints a commit message with the required Co-authored-by trailer appended.
Use --check to verify an existing message already includes the trailer.
EOF
}

has_trailer() {
  local message="$1"
  grep -Fq "${AGENT_CO_AUTHOR_TRAILER}" <<<"${message}"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ "${1:-}" == "--check" ]]; then
  message="$(cat)"
  if has_trailer "${message}"; then
    exit 0
  fi
  echo "missing required trailer: ${AGENT_CO_AUTHOR_TRAILER}" >&2
  exit 1
fi

if [[ $# -lt 1 ]]; then
  usage >&2
  exit 2
fi

subject="$1"
body="${2:-}"

if [[ -n "${body}" ]]; then
  printf '%s\n\n%s\n%s\n' "${subject}" "${body}" "${AGENT_CO_AUTHOR_TRAILER}"
else
  printf '%s\n\n%s\n' "${subject}" "${AGENT_CO_AUTHOR_TRAILER}"
fi
