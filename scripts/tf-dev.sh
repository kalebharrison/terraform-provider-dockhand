#!/usr/bin/env bash
set -euo pipefail

# Local developer workflow for this provider without publishing to a registry.
#
# Usage:
#   /path/to/terraform-provider-dockhand/scripts/tf-dev.sh plan
#   /path/to/terraform-provider-dockhand/scripts/tf-dev.sh apply
#   /path/to/terraform-provider-dockhand/scripts/tf-dev.sh --chdir /path/to/tf apply
#
# Notes:
# - This uses a Terraform CLI config with dev_overrides, and then runs terraform.
# - Terraform may warn that "terraform init" isn't necessary with dev overrides. That's expected.

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bin_dir="${repo_root}/bin"

workdir="${PWD}"
if [[ "${1:-}" == "--chdir" ]]; then
  if [[ -z "${2:-}" ]]; then
    echo "missing value for --chdir" >&2
    exit 2
  fi
  workdir="${2}"
  shift 2
fi

cli_cfg="${workdir}/.terraformrc.dockhand-dev"

if [[ ! -x "${bin_dir}/terraform-provider-dockhand" ]]; then
  echo "Provider binary not found at ${bin_dir}/terraform-provider-dockhand" >&2
  echo "Run: (cd ${repo_root} && make tf-dev-build)" >&2
  exit 2
fi

mkdir -p "${workdir}"

cat > "${cli_cfg}" <<HCL
provider_installation {
  dev_overrides {
    "kalebharrison/dockhand" = "${bin_dir}"
  }
  direct {}
}
HCL

cd "${workdir}"
export TF_CLI_CONFIG_FILE="${cli_cfg}"

if command -v tofu >/dev/null 2>&1; then
  exec tofu "$@"
fi

exec terraform "$@"
