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
  icon            = "globe"

  collect_activity         = true
  collect_metrics          = true
  highlight_changes        = true
  update_check_enabled     = false
  update_check_auto_update = false
  image_prune_enabled      = false
  timezone                 = "UTC"
}
```

## Notes

- `socket_path` is required when `connection_type = "socket"`.
- Additional environment fields returned by Dockhand that are not yet mapped are intentionally omitted.

