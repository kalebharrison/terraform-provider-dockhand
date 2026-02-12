# dockhand_schedules (Data Source)

Reads schedules from `/api/schedules`.

## Example Usage

```terraform
data "dockhand_schedules" "this" {}
```

## Schema

### Read-Only

- `id` (String) Static ID: `dockhand-schedules`.
- `schedules` (List of Object)
  - `id` (String)
  - `type` (String)
  - `name` (String)
  - `entity_name` (String)
  - `description` (String)
  - `environment_id` (String)
  - `environment_name` (String)
  - `enabled` (Bool)
  - `schedule_type` (String)
  - `cron_expression` (String)
  - `next_run` (String)
  - `is_system` (Bool)
  - `last_status` (String)
  - `last_triggered_at` (String)
  - `last_completed_at` (String)

## Notes

- Current API contract observed in this environment is read-only for schedules (`GET /api/schedules`).
- System cleanup schedules are currently managed through general settings and internal Dockhand behavior.

