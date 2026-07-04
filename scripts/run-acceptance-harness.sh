#!/usr/bin/env bash
# shellcheck shell=bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

DOCKHAND_IMAGE="${DOCKHAND_IMAGE:-fnsys/dockhand:latest}"
DIND_IMAGE="${DIND_IMAGE:-docker:27-dind}"
HAWSER_IMAGE="${HAWSER_IMAGE:-ghcr.io/finsys/hawser:latest}"
REGISTRY_IMAGE="${REGISTRY_IMAGE:-registry:2}"
TEST_REGEX="${TEST_REGEX:-TestAcc}"
export TF_ACC="${TF_ACC:-1}"
RUN_ENDPOINT_PROBE="${RUN_ENDPOINT_PROBE:-true}"
RUN_WEBUI_AUDIT="${RUN_WEBUI_AUDIT:-false}"
RUN_DOCS_REFERENCE_AUDIT="${RUN_DOCS_REFERENCE_AUDIT:-false}"
RUN_PRIVATE_ENDPOINT_PROBE="${RUN_PRIVATE_ENDPOINT_PROBE:-false}"
DOCKHAND_TEST_ENDPOINT="${DOCKHAND_TEST_ENDPOINT:-http://127.0.0.1:13001}"
DOCKHAND_TEST_USERNAME="${DOCKHAND_TEST_USERNAME:-tfacc}"
# Default matches acceptance-ci.yml; ephemeral CI-only creds, not production secrets.
DOCKHAND_TEST_PASSWORD="${DOCKHAND_TEST_PASSWORD:-tfaccpass123!}"
DOCKHAND_TEST_AUTH_PROVIDER="${DOCKHAND_TEST_AUTH_PROVIDER:-local}"
DOCKHAND_TEST_DIND_HOST="${DOCKHAND_TEST_DIND_HOST:-}"
DOCKHAND_TEST_DIND_PORT="${DOCKHAND_TEST_DIND_PORT:-2375}"
DOCKHAND_TEST_ENABLE_PRUNE_ACTIONS="${DOCKHAND_TEST_ENABLE_PRUNE_ACTIONS:-true}"
DOCKHAND_TEST_PRUNE_MODE="${DOCKHAND_TEST_PRUNE_MODE:-containers}"
DOCKHAND_TEST_REGISTRY_URL="${DOCKHAND_TEST_REGISTRY_URL:-http://registry:5000}"
DOCKHAND_TEST_REGISTRY_HOST_PORT="${DOCKHAND_TEST_REGISTRY_HOST_PORT:-25000}"
DOCKHAND_TEST_GIT_HELPER_REPO_URL="${DOCKHAND_TEST_GIT_HELPER_REPO_URL:-https://github.com/docker/awesome-compose.git}"
DOCKHAND_TEST_GIT_HELPER_BRANCH="${DOCKHAND_TEST_GIT_HELPER_BRANCH:-master}"
DOCKHAND_TEST_GIT_HELPER_COMPOSE_PATH="${DOCKHAND_TEST_GIT_HELPER_COMPOSE_PATH:-nginx-flask-mysql/compose.yaml}"
DOCKHAND_TEST_GIT_STACK_REPO_URL="${DOCKHAND_TEST_GIT_STACK_REPO_URL:-${DOCKHAND_TEST_GIT_HELPER_REPO_URL}}"
DOCKHAND_TEST_GIT_STACK_BRANCH="${DOCKHAND_TEST_GIT_STACK_BRANCH:-${DOCKHAND_TEST_GIT_HELPER_BRANCH}}"
# postgresql-pgadmin ships a committed .env beside compose.yaml (nginx-flask-mysql does not).
DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH="${DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH:-postgresql-pgadmin/compose.yaml}"
TF_ACC_TERRAFORM_PATH="${TF_ACC_TERRAFORM_PATH:-}"
TF_ACC_TERRAFORM_VERSION="${TF_ACC_TERRAFORM_VERSION:-1.14.8}"

if [[ -z "${TF_ACC_TERRAFORM_PATH}" ]]; then
  if command -v terraform >/dev/null 2>&1; then
    TF_ACC_TERRAFORM_PATH="$(command -v terraform)"
  fi
fi

if [[ -n "${TF_ACC_TERRAFORM_PATH}" ]]; then
  export TF_ACC_TERRAFORM_PATH
fi

if [[ -n "${TF_ACC_TERRAFORM_VERSION}" ]]; then
  export TF_ACC_TERRAFORM_VERSION
fi

SUFFIX="$(date +%s)"
NETWORK_NAME="dockhand-ci-${SUFFIX}"
DIND_CONTAINER="dind-${SUFFIX}"
DOCKHAND_CONTAINER="dockhand-${SUFFIX}"
HAWSER_CONTAINER="hawser-${SUFFIX}"
REGISTRY_CONTAINER="registry-${SUFFIX}"

if [[ -z "${DOCKHAND_TEST_DIND_HOST}" ]]; then
  DOCKHAND_TEST_DIND_HOST="${DIND_CONTAINER}"
fi

dump_logs() {
  echo "Dockhand logs:"
  docker logs "${DOCKHAND_CONTAINER}" || true
  echo "DinD logs:"
  docker logs "${DIND_CONTAINER}" || true
  echo "Registry logs:"
  docker logs "${REGISTRY_CONTAINER}" || true
  echo "Hawser logs:"
  docker --host "tcp://127.0.0.1:23750" logs "${HAWSER_CONTAINER}" 2>/dev/null || true
}

cleanup() {
  local exit_code="${1:-0}"
  if [[ "${exit_code}" -ne 0 ]]; then
    dump_logs
  fi
  docker --host "tcp://127.0.0.1:23750" rm -f "${HAWSER_CONTAINER}" >/dev/null 2>&1 || true
  docker rm -f "${DOCKHAND_CONTAINER}" "${DIND_CONTAINER}" "${REGISTRY_CONTAINER}" >/dev/null 2>&1 || true
  docker network rm "${NETWORK_NAME}" >/dev/null 2>&1 || true
}
trap 'exit_code=$?; cleanup "${exit_code}"; exit "${exit_code}"' EXIT

login() {
  local code
  code="$(curl -sS -o /tmp/dh-login.json -w "%{http_code}" \
    -c /tmp/dh-cookies.txt \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${DOCKHAND_TEST_USERNAME}\",\"password\":\"${DOCKHAND_TEST_PASSWORD}\",\"authProvider\":\"${DOCKHAND_TEST_AUTH_PROVIDER}\"}" \
    "${DOCKHAND_TEST_ENDPOINT}/api/auth/login" || true)"
  [[ "${code}" =~ ^2[0-9][0-9]$ ]]
}

container_runtime_state() {
  local container_id="$1"
  local env_id="$2"
  curl -sS -b /tmp/dh-cookies.txt \
    "${DOCKHAND_TEST_ENDPOINT}/api/containers/${container_id}?env=${env_id}" \
    | jq -r '.state // .status // empty' 2>/dev/null || true
}

ensure_file_container_running() {
  local container_id="$1"
  local env_id="$2"
  local label="$3"
  for _ in $(seq 1 90); do
    local ctr_state
    ctr_state="$(container_runtime_state "${container_id}" "${env_id}")"
    if [[ "${ctr_state}" == "running" ]]; then
      return 0
    fi
    curl -sS -b /tmp/dh-cookies.txt -X POST \
      "${DOCKHAND_TEST_ENDPOINT}/api/containers/${container_id}/start?env=${env_id}" >/dev/null 2>&1 || true
    sleep 2
  done
  echo "Bootstrap file container ${label} (${container_id}) did not reach running state" >&2
  return 1
}

echo "Creating Docker network ${NETWORK_NAME}"
docker network create "${NETWORK_NAME}" >/dev/null

echo "Starting local registry ${REGISTRY_CONTAINER}"
docker run -d --name "${REGISTRY_CONTAINER}" --network "${NETWORK_NAME}" \
  --network-alias registry \
  -p "${DOCKHAND_TEST_REGISTRY_HOST_PORT}:5000" \
  "${REGISTRY_IMAGE}" >/dev/null

echo "Starting DinD ${DIND_CONTAINER}"
docker run -d --name "${DIND_CONTAINER}" --network "${NETWORK_NAME}" --privileged \
  --network-alias "${DIND_CONTAINER}" \
  --network-alias dind \
  -e DOCKER_TLS_CERTDIR= \
  -p 23750:2375 \
  "${DIND_IMAGE}" \
  --insecure-registry=registry:5000 \
  --tls=false \
  --host=tcp://0.0.0.0:2375 \
  --host=unix:///var/run/docker.sock >/dev/null

echo "Waiting for DinD API"
for _ in $(seq 1 90); do
  if docker --host "tcp://127.0.0.1:23750" version >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
docker --host "tcp://127.0.0.1:23750" version >/dev/null

echo "Starting Dockhand ${DOCKHAND_CONTAINER} with image ${DOCKHAND_IMAGE}"
docker run -d --name "${DOCKHAND_CONTAINER}" --network "${NETWORK_NAME}" \
  -p 13001:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  "${DOCKHAND_IMAGE}" >/dev/null

echo "Waiting for Dockhand API"
for _ in $(seq 1 60); do
  code="$(curl -sS -o /dev/null -w "%{http_code}" "${DOCKHAND_TEST_ENDPOINT}/api/auth/session" || true)"
  if [[ "${code}" != "000" && "${code}" != "502" && "${code}" != "503" ]]; then
    break
  fi
  sleep 2
done

if ! login; then
  echo "Bootstrapping first admin"
  user_payload="$(jq -nc --arg u "${DOCKHAND_TEST_USERNAME}" --arg p "${DOCKHAND_TEST_PASSWORD}" \
    '{username:$u,password:$p,email:"tfacc@example.local",displayName:"TF Acc CI",isAdmin:true,isActive:true}')"
  curl -sS -o /tmp/dh-bootstrap-user.json -w "%{http_code}\n" \
    -H "Content-Type: application/json" \
    -d "${user_payload}" \
    "${DOCKHAND_TEST_ENDPOINT}/api/users" >/tmp/dh-bootstrap-user.code || true

  auth_payload='{"authEnabled":true,"defaultProvider":"local","sessionTimeout":86400}'
  curl -sS -o /tmp/dh-bootstrap-auth.json -w "%{http_code}\n" \
    -X PUT \
    -H "Content-Type: application/json" \
    -d "${auth_payload}" \
    "${DOCKHAND_TEST_ENDPOINT}/api/auth/settings" >/tmp/dh-bootstrap-auth.code || true

  for _ in $(seq 1 20); do
    if login; then
      break
    fi
    sleep 2
  done
fi
login

existing_id="$(
  curl -sS -b /tmp/dh-cookies.txt "${DOCKHAND_TEST_ENDPOINT}/api/environments" \
    | jq -r '.[] | select(.name=="ci-dind") | .id' \
    | head -n 1
)"
if [[ -z "${existing_id}" ]]; then
  payload="$(jq -nc --arg host "${DOCKHAND_TEST_DIND_HOST}" --argjson port "${DOCKHAND_TEST_DIND_PORT}" \
    '{name:"ci-dind",connectionType:"direct",host:$host,port:$port,protocol:"http",tlsSkipVerify:false,collectActivity:true,collectMetrics:true,highlightChanges:true,icon:"globe"}')"
  curl -sS -b /tmp/dh-cookies.txt \
    -H "Content-Type: application/json" \
    -d "${payload}" \
    "${DOCKHAND_TEST_ENDPOINT}/api/environments" >/tmp/dh-created-env.json
  existing_id="$(jq -r '.id // empty' /tmp/dh-created-env.json)"
fi
if [[ -z "${existing_id}" ]]; then
  echo "Failed to resolve ci-dind environment id" >&2
  exit 1
fi

stack_adopt_name="tf-acc-adopt-${SUFFIX}"
curl -sS -b /tmp/dh-cookies.txt \
  "${DOCKHAND_TEST_ENDPOINT}/api/stacks/default-path?name=${stack_adopt_name}" >/tmp/dh-stack-adopt-path.json
stack_adopt_compose_path="$(jq -r '.composePath // empty' /tmp/dh-stack-adopt-path.json)"
if [[ -z "${stack_adopt_compose_path}" ]]; then
  echo "Failed to resolve stack adopt compose path for ${stack_adopt_name}" >&2
  cat /tmp/dh-stack-adopt-path.json >&2 || true
  exit 1
fi
stack_adopt_dir="$(dirname "${stack_adopt_compose_path}")"
cat > /tmp/dh-stack-adopt-compose.yaml <<EOF
services:
  app:
    image: busybox:1.36.1
    command:
      - sh
      - -c
      - sleep 3600
EOF
docker exec "${DOCKHAND_CONTAINER}" mkdir -p "${stack_adopt_dir}"
docker cp /tmp/dh-stack-adopt-compose.yaml "${DOCKHAND_CONTAINER}:${stack_adopt_compose_path}"

echo "Bootstrapping acceptance fixtures"

bootstrap_ctr="tf-acc-bootstrap-file-${SUFFIX}"
create_ctr_payload="$(jq -nc --arg n "${bootstrap_ctr}" '{name:$n,image:"busybox:1.36.1",command:["sleep","3600"],enabled:true}')"
curl -sS -b /tmp/dh-cookies.txt -H "Content-Type: application/json" \
  -d "${create_ctr_payload}" \
  "${DOCKHAND_TEST_ENDPOINT}/api/containers?env=${existing_id}" >/tmp/dh-bootstrap-container.json
file_container_id="$(jq -r '.id // empty' /tmp/dh-bootstrap-container.json)"
if [[ -z "${file_container_id}" ]]; then
  echo "Failed to bootstrap file container fixture" >&2
  cat /tmp/dh-bootstrap-container.json >&2 || true
  exit 1
fi
curl -sS -b /tmp/dh-cookies.txt -X POST \
  "${DOCKHAND_TEST_ENDPOINT}/api/containers/${file_container_id}/start?env=${existing_id}" >/tmp/dh-bootstrap-container-start.json || true
ensure_file_container_running "${file_container_id}" "${existing_id}" "${bootstrap_ctr}"
export DOCKHAND_TEST_FILE_CONTAINER_ID="${file_container_id}"
export DOCKHAND_TEST_FILE_CONTAINER_ENV_ID="${existing_id}"

registry_payload="$(jq -nc --arg url "http://registry:5000" --arg name "ci-catalog-${SUFFIX}" \
  '{name:$name,url:$url,isDefault:false}')"
curl -sS -b /tmp/dh-cookies.txt -H "Content-Type: application/json" \
  -d "${registry_payload}" \
  "${DOCKHAND_TEST_ENDPOINT}/api/registries" >/tmp/dh-bootstrap-registry.json
registry_catalog_id="$(jq -r '.id // empty' /tmp/dh-bootstrap-registry.json)"
if [[ -n "${registry_catalog_id}" ]]; then
  export DOCKHAND_TEST_REGISTRY_CATALOG_ID="${registry_catalog_id}"
fi

git_repo_payload="$(jq -nc \
  --arg name "ci-git-repo-${SUFFIX}" \
  --arg url "${DOCKHAND_TEST_GIT_STACK_REPO_URL}" \
  --arg branch "${DOCKHAND_TEST_GIT_STACK_BRANCH}" \
  --arg compose "${DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH}" \
  --argjson env_id "${existing_id}" \
  '{name:$name,url:$url,branch:$branch,composePath:$compose,environmentId:$env_id}')"
curl -sS -b /tmp/dh-cookies.txt -H "Content-Type: application/json" \
  -d "${git_repo_payload}" \
  "${DOCKHAND_TEST_ENDPOINT}/api/git/repositories" >/tmp/dh-bootstrap-git-repo.json
git_repo_id="$(jq -r '.id // empty' /tmp/dh-bootstrap-git-repo.json)"
if [[ -n "${git_repo_id}" ]]; then
  git_stack_name="tf-acc-bootstrap-git-${SUFFIX}"
  git_stack_payload="$(jq -nc \
    --arg name "${git_stack_name}" \
    --argjson env_id "${existing_id}" \
    --argjson repo_id "${git_repo_id}" \
    --arg compose "${DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH}" \
    '{stackName:$name,environmentId:$env_id,repositoryId:$repo_id,composePath:$compose,deployNow:false,buildOnDeploy:false,repullImages:false,autoUpdateEnabled:false,autoUpdateCron:"0 3 * * *",webhookEnabled:false}')"
  curl -sS -b /tmp/dh-cookies.txt -H "Content-Type: application/json" \
    -d "${git_stack_payload}" \
    "${DOCKHAND_TEST_ENDPOINT}/api/git/stacks?env=${existing_id}" >/tmp/dh-bootstrap-git-stack.json
  git_stack_id="$(jq -r '.id // empty' /tmp/dh-bootstrap-git-stack.json)"
  if [[ -n "${git_stack_id}" ]]; then
    export DOCKHAND_TEST_GIT_STACK_ID="${git_stack_id}"
    env_path=""
    for _ in $(seq 1 45); do
      curl -sS -b /tmp/dh-cookies.txt \
        "${DOCKHAND_TEST_ENDPOINT}/api/git/stacks/${git_stack_id}/env-files" >/tmp/dh-env-files.json || true
      env_path="$(jq -r '.files[]? | select(. == ".env") // empty' /tmp/dh-env-files.json | head -n 1)"
      if [[ -z "${env_path}" ]]; then
        env_path="$(jq -r '.files[0] // empty' /tmp/dh-env-files.json)"
      fi
      if [[ -n "${env_path}" ]]; then
        break
      fi
      sleep 2
    done
    if [[ -n "${env_path}" ]]; then
      export DOCKHAND_TEST_GIT_STACK_ENV_PATH="${env_path}"
    fi
  fi
fi

agent_token="ci-agent-${SUFFIX}"
dockhand_ip="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${DOCKHAND_CONTAINER}")"
dockhand_ws_url="ws://${dockhand_ip}:3000/api/hawser/connect"

export TF_ACC=1
export DOCKHAND_TEST_ENDPOINT="${DOCKHAND_TEST_ENDPOINT}"
export DOCKHAND_TEST_USERNAME="${DOCKHAND_TEST_USERNAME}"
export DOCKHAND_TEST_PASSWORD="${DOCKHAND_TEST_PASSWORD}"
export DOCKHAND_TEST_AUTH_PROVIDER="${DOCKHAND_TEST_AUTH_PROVIDER}"
export DOCKHAND_TEST_DIND_HOST="${DOCKHAND_TEST_DIND_HOST}"
export DOCKHAND_TEST_DIND_PORT="${DOCKHAND_TEST_DIND_PORT}"
export DOCKHAND_TEST_ENABLE_PRUNE_ACTIONS="${DOCKHAND_TEST_ENABLE_PRUNE_ACTIONS}"
export DOCKHAND_TEST_PRUNE_MODE="${DOCKHAND_TEST_PRUNE_MODE}"
export DOCKHAND_TEST_DEFAULT_ENV="${existing_id}"
export DOCKHAND_TEST_REGISTRY_URL="${DOCKHAND_TEST_REGISTRY_URL}"
export DOCKHAND_TEST_REGISTRY_HOST_URL="http://127.0.0.1:${DOCKHAND_TEST_REGISTRY_HOST_PORT}"
export DOCKHAND_TEST_GIT_HELPER_REPO_URL="${DOCKHAND_TEST_GIT_HELPER_REPO_URL}"
export DOCKHAND_TEST_GIT_HELPER_BRANCH="${DOCKHAND_TEST_GIT_HELPER_BRANCH}"
export DOCKHAND_TEST_GIT_HELPER_COMPOSE_PATH="${DOCKHAND_TEST_GIT_HELPER_COMPOSE_PATH}"
export DOCKHAND_TEST_GIT_STACK_REPO_URL="${DOCKHAND_TEST_GIT_STACK_REPO_URL:-${DOCKHAND_TEST_GIT_HELPER_REPO_URL}}"
export DOCKHAND_TEST_GIT_STACK_BRANCH="${DOCKHAND_TEST_GIT_STACK_BRANCH:-${DOCKHAND_TEST_GIT_HELPER_BRANCH}}"
export DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH="${DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH:-${DOCKHAND_TEST_GIT_HELPER_COMPOSE_PATH}}"
export DOCKHAND_TEST_GIT_REPO_ENV_ID="${DOCKHAND_TEST_GIT_REPO_ENV_ID:-${existing_id}}"
export DOCKHAND_TEST_STACK_ADOPT_ENV_ID="${existing_id}"
export DOCKHAND_TEST_STACK_ADOPT_NAME="${stack_adopt_name}"
export DOCKHAND_TEST_STACK_ADOPT_COMPOSE_PATH="${stack_adopt_compose_path}"
export DOCKHAND_TEST_AGENT_TOKEN="${agent_token}"
export DOCKHAND_TEST_AGENT_NAME="ci-hawser"
export DOCKHAND_TEST_HAWSER_SERVER_URL="${dockhand_ws_url}"
export DOCKHAND_TEST_HAWSER_CONTAINER="${HAWSER_CONTAINER}"
export DOCKHAND_TEST_HAWSER_DOCKER_HOST="tcp://127.0.0.1:23750"
export DOCKHAND_TEST_HAWSER_IMAGE="${HAWSER_IMAGE}"

export DOCKHAND_ENDPOINT="${DOCKHAND_TEST_ENDPOINT}"
export DOCKHAND_USERNAME="${DOCKHAND_TEST_USERNAME}"
export DOCKHAND_PASSWORD="${DOCKHAND_TEST_PASSWORD}"
export DOCKHAND_AUTH_PROVIDER="${DOCKHAND_TEST_AUTH_PROVIDER}"
export DOCKHAND_DEFAULT_ENV="${existing_id}"
export DOCKHAND_INSECURE="${DOCKHAND_INSECURE:-false}"

ensure_file_container_running "${file_container_id}" "${existing_id}" "${bootstrap_ctr}"

if [[ -n "${DOCKHAND_TEST_HAWSER_CONTAINER:-}" ]]; then
  hawser_host="${DOCKHAND_TEST_HAWSER_DOCKER_HOST:-tcp://127.0.0.1:23750}"
  docker --host "${hawser_host}" rm -f "${DOCKHAND_TEST_HAWSER_CONTAINER}" >/dev/null 2>&1 || true
fi

echo "Running acceptance tests with regex: ${TEST_REGEX}"
TEST_JSON="/tmp/acceptance-test-${SUFFIX}.jsonl"
set +e
GOCACHE="${GOCACHE:-$PWD/.cache/go-build}" \
GOMODCACHE="${GOMODCACHE:-$PWD/.cache/gomod}" \
go test -json ./internal/provider -run "${TEST_REGEX}" -timeout 45m | tee "${TEST_JSON}"
test_exit="${PIPESTATUS[0]}"
set -e
/usr/bin/python3 scripts/check-acceptance-skips.py "${TEST_JSON}"
if [[ "${test_exit}" -ne 0 ]]; then
  exit "${test_exit}"
fi

ensure_file_container_running "${file_container_id}" "${existing_id}" "${bootstrap_ctr}"

if [[ "${RUN_ENDPOINT_PROBE}" == "true" ]]; then
  echo "Running endpoint compatibility probe"
  DOCKHAND_PROBE_ALLOW_MUTATION=false /usr/bin/python3 scripts/endpoint-probe.py
fi

if [[ "${RUN_WEBUI_AUDIT}" == "true" ]]; then
  echo "Running WebUI endpoint audit"
  DOCKHAND_ENDPOINT="${DOCKHAND_TEST_ENDPOINT}" /usr/bin/python3 scripts/webui-endpoint-audit.py
fi

if [[ "${RUN_DOCS_REFERENCE_AUDIT}" == "true" ]]; then
  echo "Running docs-reference endpoint audit"
  /usr/bin/python3 scripts/docs-reference-audit.py
fi

if [[ "${RUN_PRIVATE_ENDPOINT_PROBE}" == "true" ]]; then
  echo "Running private endpoint probe"
  /usr/bin/python3 scripts/private-endpoint-probe.py
fi

echo "Acceptance harness completed successfully"
