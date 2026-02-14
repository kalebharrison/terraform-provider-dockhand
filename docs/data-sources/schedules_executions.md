# dockhand_schedules_executions (Data Source)

Reads schedule execution history from `/api/schedules/executions`.

## Example Usage

```terraform
data "dockhand_schedules_executions" "recent" {
  limit  = 25
  offset = 0
}
```

## Schema

### Optional

- `limit` (Number) Maximum number of execution rows requested.
- `offset` (Number) Offset into execution rows.

### Read-Only

- `id` (String) Synthetic ID: `dockhand-schedule-executions:<limit>:<offset>`.
- `total` (Number) Total rows available server-side.
- `returned_limit` (Number) Effective limit returned by Dockhand.
- `returned_offset` (Number) Effective offset returned by Dockhand.
- `executions` (List of Object)
  - `id` (String)
  - `schedule_type` (String)
  - `schedule_id` (String)
  - `environment_id` (String)
  - `entity_name` (String)
  - `triggered_by` (String)
  - `triggered_at` (String)
  - `started_at` (String)
  - `completed_at` (String)
  - `duration_ms` (Number)
  - `status` (String)
  - `error_message` (String)
  - `details_json` (String)
  - `created_at` (String)
  - `logs` (String)
