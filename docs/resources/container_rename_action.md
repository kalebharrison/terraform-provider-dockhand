# dockhand_container_rename_action (Resource)

Runs a one-shot container rename.

## Example Usage

```terraform
resource "dockhand_container_rename_action" "rename" {
  env          = "2"
  container_id = "abc123..."
  name         = "new-container-name"
  trigger      = "manual-1"
}
```

## Schema

### Required

- `container_id` (String)
- `name` (String) New container name.

### Optional

- `env` (String)
- `trigger` (String)

### Read-Only

- `id` (String)
