# dockhand_container_file (Resource)

Manages a file or directory inside a running container.

## Example Usage

```terraform
resource "dockhand_container_file" "motd" {
  env          = "2"
  container_id = "abc123"
  path         = "/tmp/motd.txt"
  type         = "file"
  content      = "managed by terraform"
}
```

## Schema

### Required

- `container_id` (String)
- `path` (String)

### Optional

- `env` (String)
- `type` (String) `file` (default) or `directory`.
- `content` (String) File content. Applies when `type = "file"`.

### Read-Only

- `id` (String)
