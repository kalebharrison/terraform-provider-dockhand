# dockhand_network_connection_action (Resource)

Runs a one-shot network connect/disconnect action for a container.

## Example Usage

```terraform
resource "dockhand_network_connection_action" "attach" {
  network_id   = "2f4e9f3d20f6"
  container_id = "a1b2c3d4e5f6"
  action       = "connect"
  trigger      = "2026-02-14T10:00:00Z"
}
```

## Schema

### Required

- `network_id` (String) Network ID.
- `container_id` (String) Container ID.
- `action` (String) One of: `connect`, `disconnect`.

### Optional

- `env` (String) Optional environment ID query parameter.
- `trigger` (String) Arbitrary value; change it to re-run the action.

### Read-Only

- `id` (String) Internal action execution ID.
