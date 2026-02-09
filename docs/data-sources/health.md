# dockhand_health (Data Source)

Checks API availability by calling `/api/dashboard/stats`.

## Example Usage

```terraform
data "dockhand_health" "current" {
  env = "1"
}
```

## Schema

### Read-Only

- `id` (String) Static ID: `dockhand-health`.
- `status` (String) `ok` when the API request succeeds.
- `version` (String) Reserved for future version detection.
- `checked_at` (String) Time the request was executed.

### Optional

- `env` (String) Environment ID used for the check.
