#!/usr/bin/env bash
set -euo pipefail

# Build versioned provider zip packages for multiple OS/arch combinations.
# Output artifacts are compatible with Terraform/OpenTofu filesystem mirrors.

usage() {
  cat <<'TXT' >&2
Usage: build-packages.sh --version X.Y.Z [--out-dir DIR]

Outputs (in out-dir, default: ./dist-local):
  terraform-provider-dockhand_<version>_<os>_<arch>.zip
  terraform-provider-dockhand_<version>_SHA256SUMS
TXT
}

version=""
out_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      version="${2:-}"
      shift 2
      ;;
    --out-dir)
      out_dir="${2:-}"
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

if [[ -z "${version}" ]]; then
  echo "--version is required" >&2
  usage
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_dir="${out_dir:-${repo_root}/dist-local}"
tmp_dir="${out_dir}/.tmp"

if ! command -v zip >/dev/null 2>&1; then
  echo "zip is required (install zip)" >&2
  exit 2
fi

mkdir -p "${out_dir}" "${tmp_dir}"
rm -rf "${tmp_dir:?}/"*

platforms=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/arm64"
  "linux/amd64"
  "windows/arm64"
  "windows/amd64"
)

sum_file="${out_dir}/terraform-provider-dockhand_${version}_SHA256SUMS"
rm -f "${sum_file}"

for p in "${platforms[@]}"; do
  goos="${p%/*}"
  goarch="${p#*/}"

  exe="terraform-provider-dockhand_v${version}"
  if [[ "${goos}" == "windows" ]]; then
    exe="${exe}.exe"
  fi

  build_dir="${tmp_dir}/${goos}_${goarch}"
  mkdir -p "${build_dir}"

  (
    cd "${repo_root}"
    GOCACHE="${repo_root}/.cache/go-build" GOMODCACHE="${repo_root}/.cache/gomod" \
      CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
      go build -trimpath -ldflags "-s -w -X main.version=v${version}" -o "${build_dir}/${exe}" .
  )

  zip_name="terraform-provider-dockhand_${version}_${goos}_${goarch}.zip"
  (
    cd "${build_dir}"
    zip -q "${out_dir}/${zip_name}" "${exe}"
  )

  # Checksums
  if command -v shasum >/dev/null 2>&1; then
    (cd "${out_dir}" && shasum -a 256 "${zip_name}" >> "${sum_file}")
  else
    (cd "${out_dir}" && sha256sum "${zip_name}" >> "${sum_file}")
  fi
done

echo "Wrote packages to: ${out_dir}" >&2

