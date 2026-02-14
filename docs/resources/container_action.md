# dockhand_container_action (Resource)

Runs a one-shot action on a container.

## Example Usage

```terraform
resource "dockhand_container_action" "restart_web" {
  env          = "2"
  container_id = "abc123..."
  action       = "restart"
  trigger      = "2026-02-12T03:00:00Z"
}
```

## Schema

### Required

- `container_id` (String) Container ID to act on.
- `action` (String) Action to execute: `start`, `stop`, `restart`, `pause`, or `unpause`.

### Optional

- `env` (String) Optional environment ID query parameter.
- `trigger` (String) Arbitrary value; change it to re-run the action.

### Read-Only

- `id` (String) Internal action execution ID.
