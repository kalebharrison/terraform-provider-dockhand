# dockhand_schedule_settings (Data Source)

Reads global schedule settings via `GET /api/schedules/settings`.

## Example Usage

```hcl
data "dockhand_schedule_settings" "global" {}
```

## Schema

### Read-Only

- `id` (String) Synthetic ID (`schedule-settings`).
- `hide_system_jobs` (Boolean) Current Dockhand schedule setting value.
