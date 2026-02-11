#!/usr/bin/env bash
set -euo pipefail

# Wrapper to run Terraform/OpenTofu against the repo-local, gitignored workspace in ./terraform/dockhand.
#
# This preserves the old workflow you had under /Users/kharrison/Documents/terraform/dockhand, but keeps it
# inside the repo so Codex can read/write it without extra sandbox prompts.
#
# Usage:
#   ./scripts/tf-local.sh init
#   ./scripts/tf-local.sh plan
#   ./scripts/tf-local.sh apply
#   ./scripts/tf-local.sh destroy
#
# Optional:
#   ./scripts/tf-local.sh --chdir ./terraform/dockhand/test plan

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

workdir="${repo_root}/terraform/dockhand/test"
if [[ "${1:-}" == "--chdir" ]]; then
  workdir="${2:-}"
  shift 2
fi

cfg_dir="${repo_root}/terraform/dockhand"
tf_cfg="${cfg_dir}/terraformrc.dockhand"
tofu_cfg="${cfg_dir}/tofurc.dockhand"

if [[ ! -d "${cfg_dir}" ]]; then
  echo "Expected repo-local workspace at: ${cfg_dir}" >&2
  echo "If it doesn't exist, create it (or copy from your previous local folder)." >&2
  exit 2
fi

if [[ -f "${workdir}/env.sh" ]]; then
  # shellcheck disable=SC1090
  source "${workdir}/env.sh"
fi

if [[ -f "${tf_cfg}" ]]; then
  export TF_CLI_CONFIG_FILE="${tf_cfg}"
fi
if [[ -f "${tofu_cfg}" ]]; then
  export TOFU_CLI_CONFIG_FILE="${tofu_cfg}"
fi

cd "${workdir}"

if command -v tofu >/dev/null 2>&1; then
  exec tofu "$@"
fi

exec terraform "$@"

