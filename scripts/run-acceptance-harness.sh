#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

DOCKHAND_IMAGE="${DOCKHAND_IMAGE:-fnsys/dockhand:latest}"
DIND_IMAGE="${DIND_IMAGE:-docker:27-dind}"
HAWSER_IMAGE="${HAWSER_IMAGE:-ghcr.io/finsys/hawser:latest}"
TEST_REGEX="${TEST_REGEX:-TestAcc}"
RUN_ENDPOINT_PROBE="${RUN_ENDPOINT_PROBE:-true}"
RUN_WEBUI_AUDIT="${RUN_WEBUI_AUDIT:-false}"
RUN_DOCS_REFERENCE_AUDIT="${RUN_DOCS_REFERENCE_AUDIT:-false}"
RUN_PRIVATE_ENDPOINT_PROBE="${RUN_PRIVATE_ENDPOINT_PROBE:-false}"
DOCKHAND_TEST_ENDPOINT="${DOCKHAND_TEST_ENDPOINT:-http://127.0.0.1:13001}"
DOCKHAND_TEST_USERNAME="${DOCKHAND_TEST_USERNAME:-tfacc}"
DOCKHAND_TEST_PASSWORD="${DOCKHAND_TEST_PASSWORD:-tfaccpass123!}"
DOCKHAND_TEST_AUTH_PROVIDER="${DOCKHAND_TEST_AUTH_PROVIDER:-local}"
DOCKHAND_TEST_DIND_HOST="${DOCKHAND_TEST_DIND_HOST:-}"
DOCKHAND_TEST_DIND_PORT="${DOCKHAND_TEST_DIND_PORT:-2375}"
DOCKHAND_TEST_ENABLE_PRUNE_ACTIONS="${DOCKHAND_TEST_ENABLE_PRUNE_ACTIONS:-true}"
DOCKHAND_TEST_PRUNE_MODE="${DOCKHAND_TEST_PRUNE_MODE:-containers}"

SUFFIX="$(date +%s)"
NETWORK_NAME="dockhand-ci-${SUFFIX}"
DIND_CONTAINER="dind-${SUFFIX}"
DOCKHAND_CONTAINER="dockhand-${SUFFIX}"
HAWSER_CONTAINER="hawser-${SUFFIX}"

if [[ -z "${DOCKHAND_TEST_DIND_HOST}" ]]; then
  DOCKHAND_TEST_DIND_HOST="${DIND_CONTAINER}"
fi

cleanup() {
  docker --host "tcp://127.0.0.1:23750" rm -f "${HAWSER_CONTAINER}" >/dev/null 2>&1 || true
  docker rm -f "${DOCKHAND_CONTAINER}" "${DIND_CONTAINER}" >/dev/null 2>&1 || true
  docker network rm "${NETWORK_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

login() {
  local code
  code="$(curl -sS -o /tmp/dh-login.json -w "%{http_code}" \
    -c /tmp/dh-cookies.txt \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${DOCKHAND_TEST_USERNAME}\",\"password\":\"${DOCKHAND_TEST_PASSWORD}\",\"authProvider\":\"${DOCKHAND_TEST_AUTH_PROVIDER}\"}" \
    "${DOCKHAND_TEST_ENDPOINT}/api/auth/login" || true)"
  [[ "${code}" =~ ^2[0-9][0-9]$ ]]
}

echo "Creating Docker network ${NETWORK_NAME}"
docker network create "${NETWORK_NAME}" >/dev/null

echo "Starting DinD ${DIND_CONTAINER}"
docker run -d --name "${DIND_CONTAINER}" --network "${NETWORK_NAME}" --privileged \
  --network-alias "${DIND_CONTAINER}" \
  --network-alias dind \
  -e DOCKER_TLS_CERTDIR= \
  -p 23750:2375 \
  "${DIND_IMAGE}" \
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

echo "Running acceptance tests with regex: ${TEST_REGEX}"
GOCACHE="${GOCACHE:-$PWD/.cache/go-build}" \
GOMODCACHE="${GOMODCACHE:-$PWD/.cache/gomod}" \
go test -v ./internal/provider -run "${TEST_REGEX}"

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
