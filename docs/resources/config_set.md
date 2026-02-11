# dockhand_config_set

Manages a Dockhand Config Set via `/api/config-sets`.

## Example Usage

```terraform
resource "dockhand_config_set" "defaults" {
  name        = "defaults"
  description = "Default container settings"

  env_vars = {
    TZ = "America/New_York"
  }

  labels = {
    "com.example.owner" = "terraform"
  }

  ports = [
    {
      container_port = 80
      host_port      = 8080
      protocol       = "tcp"
    }
  ]

  volumes = [
    {
      source    = "/tmp"
      target    = "/data"
      type      = "bind"
      read_only = true
    }
  ]

  network_mode   = "bridge"
  restart_policy = "no"
}
```

## Schema

### Read-only

- `id` (String)
- `created_at` (String)
- `updated_at` (String)

### Required

- `name` (String)

### Optional/Computed

- `description` (String)
- `env_vars` (Map of String)
- `labels` (Map of String)
- `ports` (List of Object)
- `volumes` (List of Object)
- `network_mode` (String)
- `restart_policy` (String)

