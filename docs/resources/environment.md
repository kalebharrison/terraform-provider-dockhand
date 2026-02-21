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
  timezone                 = "UTC"
}
```

## Notes

- `socket_path` is required when `connection_type = "socket"`.
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
- Known gap: environment-level vulnerability scanner selection (for example choosing Grype vs Trivy in the UI) is not yet exposed, because a stable scanner-selection API contract has not been confirmed from tested endpoints.
- Some Dockhand builds may not return cert/key bodies on read for security reasons. The provider preserves prior state values in that case.
