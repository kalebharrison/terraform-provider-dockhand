#!/usr/bin/env bash
# shellcheck shell=bash
# Agent commit helpers. Source before git commit on agent/** branches.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export AGENT_CO_AUTHOR_TRAILER="${AGENT_CO_AUTHOR_TRAILER:-Co-authored-by: Cursor Agent <noreply@cursor.com>}"

agent_commit() {
  if [[ $# -lt 1 ]]; then
    echo "usage: agent_commit \"subject\" [body]" >&2
    return 2
  fi

  local message
  message="$("${ROOT_DIR}/scripts/agent-commit-msg.sh" "$@")"
  git commit -m "${message}"
}

echo "Agent commit helpers loaded."
echo "Required trailer: ${AGENT_CO_AUTHOR_TRAILER}"
echo "Use: agent_commit \"fix(provider): summary\""
