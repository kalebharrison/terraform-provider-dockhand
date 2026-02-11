#!/usr/bin/env bash
set -euo pipefail

# Tag -> wait for GH Actions release -> download release zips -> terraform init/plan/apply/destroy.
#
# This is the preferred workflow for validating provider changes, because it tests the same
# release artifacts users will install from a mirror.
#
# Usage:
#   ./scripts/release-test.sh 0.1.4
#
# Requirements:
# - gh CLI authenticated
# - terraform installed (or tofu; terraform is used here explicitly)
# - terraform/dockhand/test/env.sh exists (gitignored) and exports DOCKHAND_* vars

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ver="${1:-}"

if [[ -z "${ver}" ]]; then
  echo "usage: release-test.sh <version>" >&2
  exit 2
fi

tag="v${ver}"

cd "${repo_root}"

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "working tree is dirty; commit first before tagging" >&2
  exit 2
fi

if git rev-parse "${tag}" >/dev/null 2>&1; then
  echo "tag already exists: ${tag}" >&2
  exit 2
fi

git tag -a "${tag}" -m "${tag}"
git push origin "${tag}"

run_id="$(gh run list --workflow release-artifacts.yml --limit 20 --json databaseId,headBranch,displayTitle | \
  /usr/bin/python3 -c "import json,sys; runs=json.load(sys.stdin); print(next((str(r['databaseId']) for r in runs if r.get('headBranch')=='${tag}'), ''))")"
if [[ -z "${run_id}" ]]; then
  echo "could not find workflow run for ${tag}; check GitHub Actions manually" >&2
  exit 2
fi

gh run watch "${run_id}" --exit-status

mirror_dir="${repo_root}/terraform/dockhand/mirror/registry.terraform.io/kalebharrison/dockhand"
mkdir -p "${mirror_dir}"

gh release download "${tag}" -R kalebharrison/terraform-provider-dockhand \
  -p "terraform-provider-dockhand_${ver}_*.zip" \
  -p "terraform-provider-dockhand_${ver}_SHA256SUMS" \
  -D "${mirror_dir}" --clobber

test_dir="${repo_root}/terraform/dockhand/test"
if [[ ! -f "${test_dir}/env.sh" ]]; then
  echo "missing ${test_dir}/env.sh (gitignored). Create it based on prior local setup." >&2
  exit 2
fi

(
  cd "${test_dir}"
  # shellcheck disable=SC1091
  source ./env.sh
  export TF_CLI_CONFIG_FILE="../terraformrc.dockhand"

  perl -pi -e "s/version\\s*=\\s*\\\"[0-9.]+\\\"/version = \\\"${ver}\\\"/g" versions.tf

  rm -rf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup
  terraform init -upgrade
  terraform plan -no-color
  terraform apply -auto-approve -no-color
  terraform destroy -auto-approve -no-color
)

