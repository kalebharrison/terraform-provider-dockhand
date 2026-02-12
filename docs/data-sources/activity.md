# dockhand_activity (Data Source)

Reads recent activity events from Dockhand.

## Example Usage

```terraform
data "dockhand_activity" "recent" {
  limit = 100
}
```

## Schema

### Optional

- `limit` (Number) Maximum number of events to return (default `50`).

### Read-Only

- `id` (String) Static ID: `dockhand-activity`.
- `events` (List of Object)
  - `id` (String)
  - `action` (String)
  - `container_id` (String)
  - `container_name` (String)
  - `image` (String)
  - `timestamp` (String)
  - `status` (String)
