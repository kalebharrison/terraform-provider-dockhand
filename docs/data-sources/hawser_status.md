# dockhand_hawser_status (Data Source)

Reads Hawser websocket endpoint status.

## Example Usage

```terraform
data "dockhand_hawser_status" "this" {}
```

## Schema

### Read-Only

- `id` (String) Static ID: `dockhand-hawser-status`.
- `status` (String) Hawser endpoint status value.
- `message` (String) Human-readable endpoint message.
- `protocol` (String) Expected WebSocket protocol URL pattern.
- `active_connections` (Number) Current active Hawser connections.
