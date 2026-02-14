# dockhand_container_pending_updates (Data Source)

Reads pending container updates from `/api/containers/pending-updates`.

## Example Usage

```terraform
data "dockhand_container_pending_updates" "this" {
  env = "2"
}
```

## Schema

### Optional

- `env` (String) Optional environment ID query parameter.

### Read-Only

- `id` (String)
- `environment_id` (String)
- `pending_updates_json` (String) Raw JSON array.
