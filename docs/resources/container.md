# dockhand_container (Resource)

Manages a Dockhand container.

## Example Usage

```terraform
resource "dockhand_container" "example" {
  name  = "tf-nginx"
  env   = "2"
  image = "nginx:alpine"

  enabled        = true
  network_mode   = "bridge"
  restart_policy = "unless-stopped"

  env_vars = {
    NGINX_ENTRYPOINT_QUIET_LOGS = "1"
  }

  labels = {
    managed_by = "terraform"
  }

  ports = [
    {
      container_port = 80
      host_port      = "18080"
      protocol       = "tcp"
    }
  ]
}
```

## Schema

### Required

- `name` (String) Container name.
- `image` (String) Container image reference.

### Optional

- `command` (String) Optional command string sent at create time.
- `enabled` (Boolean) Desired runtime state. Defaults to `true`.
- `env` (String) Optional environment ID query parameter.
- `env_vars` (Map of String) Environment variables for create request.
- `labels` (Map of String) Labels for create request.
- `network_mode` (String) Network mode for create request.
- `ports` (Attributes List) Port mappings for create request.
- `privileged` (Boolean) Create container in privileged mode.
- `restart_policy` (String) Restart policy for create request.
- `tty` (Boolean) Allocate a TTY at create time.

### Read-Only

- `id` (String) Container ID.
- `health` (String) Current health status from Dockhand.
- `restart_count` (Number) Current container restart count.
- `state` (String) Current container state.
- `status` (String) Current container status text.

## Import

Import by container ID:

```bash
terraform import dockhand_container.example <id>

# or with explicit env
terraform import dockhand_container.example <env>:<id>
```
