# dockhand_registry

Manages a Dockhand image registry via `/api/registries`.

## Example Usage

```terraform
resource "dockhand_registry" "docker_hub" {
  name       = "Docker Hub"
  url        = "https://registry.hub.docker.com"
  is_default = true
}
```

Registry with credentials:

```terraform
resource "dockhand_registry" "private" {
  name     = "My Registry"
  url      = "https://registry.example.com"
  username = "my-user"
  password = var.registry_password
}
```

To clear stored credentials:

```terraform
resource "dockhand_registry" "private" {
  name     = "My Registry"
  url      = "https://registry.example.com"
  username = ""
  password = ""
}
```

## Schema

### Read-only

- `id` (String) Registry ID.
- `has_credentials` (Boolean) Whether Dockhand has credentials stored.
- `created_at` (String)
- `updated_at` (String)

### Required

- `name` (String)
- `url` (String)

### Optional/Computed

- `is_default` (Boolean)
- `username` (String)

### Optional (Write-only)

- `password` (String, Sensitive)

