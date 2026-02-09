#!/usr/bin/env bash
set -euo pipefail

# Create a Terraform/OpenTofu filesystem mirror directory in the "packed layout"
# from locally-built zip packages.

usage() {
  cat <<'TXT' >&2
Usage: build-mirror.sh --version X.Y.Z --namespace NS [--hostname HOST] [--mirror-dir DIR] [--packages-dir DIR]

Defaults:
  --hostname     registry.terraform.io
  --mirror-dir   ./mirror
  --packages-dir ./dist-local

This writes:
  <mirror-dir>/<hostname>/<namespace>/dockhand/terraform-provider-dockhand_<version>_<os>_<arch>.zip
TXT
}

version=""
namespace=""
hostname="registry.terraform.io"
mirror_dir=""
packages_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      version="${2:-}"
      shift 2
      ;;
    --namespace)
      namespace="${2:-}"
      shift 2
      ;;
    --hostname)
      hostname="${2:-}"
      shift 2
      ;;
    --mirror-dir)
      mirror_dir="${2:-}"
      shift 2
      ;;
    --packages-dir)
      packages_dir="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown arg: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ -z "${version}" || -z "${namespace}" ]]; then
  echo "--version and --namespace are required" >&2
  usage
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mirror_dir="${mirror_dir:-${repo_root}/mirror}"
packages_dir="${packages_dir:-${repo_root}/dist-local}"

dest="${mirror_dir}/${hostname}/${namespace}/dockhand"
mkdir -p "${dest}"

shopt -s nullglob
zips=( "${packages_dir}/terraform-provider-dockhand_${version}_"*.zip )
shopt -u nullglob

if [[ ${#zips[@]} -eq 0 ]]; then
  echo "no packages found in ${packages_dir} for version ${version}" >&2
  echo "Run: ${repo_root}/scripts/build-packages.sh --version ${version}" >&2
  exit 2
fi

for z in "${zips[@]}"; do
  cp -f "${z}" "${dest}/"
done

echo "Wrote mirror to: ${mirror_dir}" >&2

