# dockhand_schedule_run_action (Resource)

Runs a one-shot schedule execution.

## Example Usage

```hcl
resource "dockhand_schedule_run_action" "run_now" {
  type        = "system_cleanup"
  schedule_id = "2"
  trigger     = "manual-run-1"
}
```

## Schema

### Required

- `type` (String) Schedule type.
- `schedule_id` (String) Dockhand schedule ID.

### Optional

- `trigger` (String) Change this value to run the action again.

### Read-Only

- `id` (String) Synthetic ID in format `<type>:<schedule_id>:<trigger>`.
