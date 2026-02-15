#!/usr/bin/env bash
set -euo pipefail

# Best-effort cleanup for known Terraform release-test fixtures.
# This prevents create-time conflicts when a prior run left fixtures behind.
#
# Requires environment variables from terraform/dockhand/test/env.sh:
# - DOCKHAND_ENDPOINT
# - DOCKHAND_USERNAME
# - DOCKHAND_PASSWORD
# - DOCKHAND_AUTH_PROVIDER (optional; defaults to local)
# - DOCKHAND_DEFAULT_ENV (optional)

endpoint="${DOCKHAND_ENDPOINT:-}"
username="${DOCKHAND_USERNAME:-}"
password="${DOCKHAND_PASSWORD:-}"
auth_provider="${DOCKHAND_AUTH_PROVIDER:-local}"
env_id="${DOCKHAND_DEFAULT_ENV:-}"

if [[ -z "${endpoint}" || -z "${username}" || -z "${password}" ]]; then
  echo "release-precheck: missing DOCKHAND_ENDPOINT/DOCKHAND_USERNAME/DOCKHAND_PASSWORD" >&2
  exit 2
fi

endpoint="${endpoint%/}"
cookie_jar="$(mktemp)"
trap 'rm -f "${cookie_jar}"' EXIT

curl -sS -f -c "${cookie_jar}" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${username}\",\"password\":\"${password}\",\"provider\":\"${auth_provider}\"}" \
  "${endpoint}/api/auth/login" >/dev/null

api_get() {
  local path="${1}"
  curl -sS -f -b "${cookie_jar}" "${endpoint}${path}"
}

api_delete() {
  local path="${1}"
  curl -sS -o /dev/null -w "%{http_code}" -b "${cookie_jar}" -X DELETE "${endpoint}${path}"
}

delete_containers_named() {
  local target_name="${1}"
  local query="/api/containers"
  if [[ -n "${env_id}" ]]; then
    query="${query}?env=${env_id}"
  fi

  local payload
  payload="$(api_get "${query}")"
  local ids
  ids="$(/usr/bin/python3 - "${target_name}" "${payload}" <<'PY'
import json
import sys
name = sys.argv[1]
raw = sys.argv[2]
try:
    items = json.loads(raw)
except Exception:
    items = []
for item in items:
    if str(item.get("name", "")) == name:
        print(item.get("id", ""))
PY
)"

  if [[ -z "${ids}" ]]; then
    return 0
  fi

  while IFS= read -r id; do
    [[ -z "${id}" ]] && continue
    local path="/api/containers/${id}?force=true"
    if [[ -n "${env_id}" ]]; then
      path="${path}&env=${env_id}"
    fi
    api_delete "${path}" >/dev/null || true
  done <<<"${ids}"
}

delete_networks_named() {
  local target_name="${1}"
  local query="/api/networks"
  if [[ -n "${env_id}" ]]; then
    query="${query}?env=${env_id}"
  fi

  local payload
  payload="$(api_get "${query}")"
  local ids
  ids="$(/usr/bin/python3 - "${target_name}" "${payload}" <<'PY'
import json
import sys
name = sys.argv[1]
raw = sys.argv[2]
try:
    items = json.loads(raw)
except Exception:
    items = []
for item in items:
    if str(item.get("name", "")) == name:
        print(item.get("id", ""))
PY
)"

  if [[ -z "${ids}" ]]; then
    return 0
  fi

  while IFS= read -r id; do
    [[ -z "${id}" ]] && continue
    local path="/api/networks/${id}"
    if [[ -n "${env_id}" ]]; then
      path="${path}?env=${env_id}"
    fi
    api_delete "${path}" >/dev/null || true
  done <<<"${ids}"
}

delete_volumes_named() {
  local target_name="${1}"
  local path="/api/volumes/${target_name}?force=true"
  if [[ -n "${env_id}" ]]; then
    path="${path}&env=${env_id}"
  fi
  api_delete "${path}" >/dev/null || true
}

delete_by_name() {
  local list_path="${1}"
  local id_key="${2}"
  local name_key="${3}"
  local target_name="${4}"
  local delete_prefix="${5}"

  local payload
  payload="$(api_get "${list_path}")"
  local ids
  ids="$(/usr/bin/python3 - "${id_key}" "${name_key}" "${target_name}" "${payload}" <<'PY'
import json
import sys
id_key, name_key, target, raw = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
try:
    items = json.loads(raw)
except Exception:
    items = []
for item in items:
    if str(item.get(name_key, "")) == target:
        print(item.get(id_key, ""))
PY
)"

  if [[ -z "${ids}" ]]; then
    return 0
  fi

  while IFS= read -r id; do
    [[ -z "${id}" ]] && continue
    api_delete "${delete_prefix}/${id}" >/dev/null || true
  done <<<"${ids}"
}

delete_test_user() {
  local payload
  if ! payload="$(api_get "/api/users" 2>/dev/null)"; then
    return 0
  fi

  local ids
  ids="$(/usr/bin/python3 - "${payload}" <<'PY'
import json
import sys
raw = sys.argv[1]
target = "tf-test-user"
try:
    items = json.loads(raw)
except Exception:
    items = []
for item in items:
    if str(item.get("username", "")) == target:
        print(item.get("id", ""))
PY
)"

  if [[ -z "${ids}" ]]; then
    return 0
  fi

  while IFS= read -r id; do
    [[ -z "${id}" ]] && continue
    api_delete "/api/users/${id}" >/dev/null || true
  done <<<"${ids}"
}

delete_containers_named "tf-action-target"
delete_networks_named "tf-test-network"
delete_volumes_named "tf-test-volume-clone"
delete_volumes_named "tf-test-volume"
delete_by_name "/api/registries" "id" "name" "Terraform Test Registry" "/api/registries"
delete_by_name "/api/git/repositories" "id" "name" "Terraform Git Repo" "/api/git/repositories"
delete_by_name "/api/git/credentials" "id" "name" "Terraform Git Credential" "/api/git/credentials"
delete_test_user

echo "release-precheck: fixture cleanup complete"
