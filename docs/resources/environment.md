# dockhand_environment

Manages a Dockhand environment via `/api/environments`.

## Example Usage

```terraform
resource "dockhand_environment" "socket" {
  name            = "truenas02"
  connection_type = "socket"
  socket_path     = "/var/run/docker.sock"

  protocol        = "http"
  port            = 2375
  tls_skip_verify = false
  # Optional mTLS inputs for TCP/TLS Docker API endpoints:
  # ca_cert     = file("${path.module}/certs/ca.pem")
  # client_cert = file("${path.module}/certs/client.pem")
  # client_key  = file("${path.module}/certs/client-key.pem")
  icon            = "globe"
  public_ip       = "203.0.113.10"

  collect_activity         = true
  collect_metrics          = true
  highlight_changes        = true
  update_check_enabled     = false
  update_check_auto_update = false
  update_check_cron        = "0 4 * * *"
  update_check_vulnerability_criteria = "never"
  image_prune_enabled      = false
  image_prune_cron         = "0 3 * * 0"
  image_prune_mode         = "dangling"
  vulnerability_scanning_enabled = true
  vulnerability_scanner          = "both" # one of: grype, trivy, both
  ensure_grype_installed         = true
  ensure_trivy_installed         = true
  timezone                 = "UTC"
}

resource "dockhand_environment" "agent" {
  name            = "edge-agent01"
  connection_type = "agent"
  agent_token     = var.edge_agent_token
}
```

## Notes

- `socket_path` is required when `connection_type = "socket"`.
- `public_ip` defaults to an empty string when unset.
- Some Dockhand builds ignore or omit `publicIp` on `POST /api/environments`. When create returns an empty/`null` `public_ip` that does not match the plan, the provider immediately follows up with `PUT /api/environments/{id}` so initial apply succeeds (same path as a later update).
- `connection_type = "agent"` maps to Dockhand `hawser-edge`.
- `agent_token` for `connection_type = "agent"` is provisioned through Dockhand `/api/hawser/tokens` with the configured raw token.
- mTLS fields are available:
  - `ca_cert`
  - `client_cert`
  - `client_key`
- Update-check scheduling fields are available:
  - `update_check_cron`
  - `update_check_vulnerability_criteria`
- Image-prune scheduling fields are available:
  - `image_prune_cron`
  - `image_prune_mode`
- Vulnerability-scanner fields are available:
  - `vulnerability_scanning_enabled`
  - `vulnerability_scanner` (`grype`, `trivy`, or `both`)
  - `ensure_grype_installed`
  - `ensure_trivy_installed`
  - computed scanner status:
    - `grype_installed`
    - `trivy_installed`
    - `grype_version`
    - `trivy_version`
- Scanner image installation uses Dockhand image pulls (`anchore/grype:latest` / `aquasec/trivy:latest`) and fails if the target environment cannot reach its Docker/image sources.
- Some Dockhand builds may not return cert/key bodies on read for security reasons. The provider preserves prior state values in that case.

## Import

```bash
terraform import dockhand_environment.example <id>
```

Numeric Dockhand environment ID.
