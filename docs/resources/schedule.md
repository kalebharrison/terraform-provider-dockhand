# dockhand_schedule

Manages pause/resume state of an existing Dockhand schedule.

## Example Usage

```hcl
resource "dockhand_schedule" "system_cleanup" {
  schedule_id = "2"
  type        = "system_cleanup"
  enabled     = true
}
```

## Notes

- This resource manages existing schedules only.
- Schedule creation is not exposed by the current API contract.
- Import format is `<type>:<schedule_id>`.

## Schema

### Required

- `schedule_id` (String) Dockhand schedule ID.
- `type` (String) Schedule type.
- `enabled` (Boolean) Desired enabled state.

### Read-Only

- `id` (String) Synthetic ID in format `<type>:<schedule_id>`.
- `name` (String) Schedule name.
- `is_system` (Boolean) Whether schedule is system-managed.
- `next_run` (String) Next run timestamp.

## Import

```bash
terraform import dockhand_schedule.system_cleanup system_cleanup:2
```
