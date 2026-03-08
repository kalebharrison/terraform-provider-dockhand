# Resource: dockhand_environment_test_action

Runs one-shot environment connectivity validation via `POST /api/environments/test`.

Change `trigger` to force Terraform to run the test again.

## Example Usage

```terraform
resource "dockhand_environment_test_action" "direct" {
  connection_type = "direct"
  host            = "dind"
  port            = 2375
  protocol        = "http"
  fail_on_error   = true
  trigger         = "run-1"
}
```

## Schema

### Required

- `connection_type` (String) Connection type (`direct`, `socket`, or `agent`).

### Optional

- `name` (String) Optional environment label sent in test payload.
- `agent_token` (String, Sensitive) Required for `connection_type = "agent"`.
- `host` (String) Required for `connection_type = "direct"`.
- `port` (Number) Required for `connection_type = "direct"`.
- `protocol` (String) Required for `connection_type = "direct"`.
- `socket_path` (String) Required for `connection_type = "socket"`.
- `tls_skip_verify` (Boolean) Skip TLS verification for HTTPS direct tests.
- `ca_cert` (String, Sensitive) PEM CA certificate (`tlsCa`).
- `client_cert` (String, Sensitive) PEM client certificate (`tlsCert`).
- `client_key` (String, Sensitive) PEM client key (`tlsKey`).
- `fail_on_error` (Boolean) Default `true`. If true, apply fails when Dockhand returns `success = false`.
- `trigger` (String) Force rerun marker.

### Read-Only

- `id` (String) Internal action instance ID (`<connection_type>:<target>:<trigger>`).
- `success` (Boolean) Dockhand connectivity test success flag.
- `error` (String) Dockhand error message when present.
- `info_json` (String) Raw `info` object JSON from Dockhand response.
- `hawser_json` (String) Raw `hawser` object JSON from Dockhand response.
- `server_version` (String) Docker server version from response `info`.
- `daemon_name` (String) Docker daemon/environment name from response `info`.
- `containers` (Number) Container count from response `info`.
- `images` (Number) Image count from response `info`.
