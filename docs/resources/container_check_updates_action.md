# dockhand_container_check_updates_action (Resource)

Runs a one-shot container update check.

## Example Usage

```terraform
resource "dockhand_container_check_updates_action" "check" {
  env     = "2"
  trigger = "manual-1"
}
```

## Schema

### Optional

- `env` (String) Optional environment ID query parameter.
- `trigger` (String) Arbitrary value; change it to re-run the action.

### Read-Only

- `id` (String) Internal action execution ID.
- `total` (Number)
- `updates_found` (Number)
- `results_json` (String) Raw JSON array of per-container results.
