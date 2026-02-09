# dockhand_settings_general

Manages Dockhand general UI/settings via `GET/POST /api/settings/general`.

This is a singleton resource. Destroying it will remove it from Terraform state but **will not** reset settings on the Dockhand server.

## Example Usage

```terraform
resource "dockhand_settings_general" "this" {
  time_format             = "24h"
  date_format             = "YYYY-MM-DD"
  show_stopped_containers = true
  highlight_updates       = true
}
```

## Schema

### Read-only

- `id` (String) Always `general`.

### Optional/Computed

All attributes are optional; omitted attributes will keep the current server value.

- `confirm_destructive` (Boolean)
- `time_format` (String)
- `date_format` (String)
- `show_stopped_containers` (Boolean)
- `highlight_updates` (Boolean)
- `default_timezone` (String)
- `download_format` (String)
- `default_grype_args` (String)
- `default_trivy_args` (String)
- `schedule_retention_days` (Number)
- `schedule_cleanup_enabled` (Boolean)
- `schedule_cleanup_cron` (String)
- `event_retention_days` (Number)
- `event_cleanup_enabled` (Boolean)
- `event_cleanup_cron` (String)
- `event_collection_mode` (String)
- `event_poll_interval` (Number)
- `metrics_collection_interval` (Number)
- `log_buffer_size_kb` (Number)
- `font` (String)
- `font_size` (String)
- `grid_font_size` (String)
- `editor_font` (String)
- `terminal_font` (String)
- `light_theme` (String)
- `dark_theme` (String)
- `primary_stack_location` (String) Set to `null` to clear.
- `external_stack_paths` (List of String)

