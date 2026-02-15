# dockhand_container_file (Resource)

Manages a text file inside a running container.

## Example Usage

```terraform
resource "dockhand_container_file" "motd" {
  env          = "2"
  container_id = "abc123"
  path         = "/tmp/motd.txt"
  content      = "managed by terraform"
}
```

## Schema

### Required

- `container_id` (String)
- `path` (String)
- `content` (String)

### Optional

- `env` (String)

### Read-Only

- `id` (String)
