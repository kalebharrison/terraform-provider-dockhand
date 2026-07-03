# Resource: dockhand_schedule_settings

Manages global schedule settings via `GET/PUT /api/schedules/settings`.

## Example Usage

```hcl
resource "dockhand_schedule_settings" "global" {
  hide_system_jobs = false
}
```

## Schema

### Required

- `hide_system_jobs` (Boolean) Whether Dockhand should hide system schedule jobs in UI views.

### Optional

- `id` (String) Singleton resource ID. Defaults to `schedule-settings`.

### Read-Only

- `id` (String) Singleton resource ID (`schedule-settings`).

## Import

```bash
terraform import dockhand_schedule_settings.example `settings`
```

Singleton resource; import ID is the literal `settings`.
