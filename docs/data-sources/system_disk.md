# dockhand_system_disk (Data Source)

Reads environment-scoped disk usage details via `GET /api/system/disk?env=<id>`.

## Example Usage

```hcl
data "dockhand_system_disk" "jetson01" {
  env = "1"
}
```

## Schema

### Optional

- `env` (String) Environment ID. If omitted, the provider default environment is used.

### Read-Only

- `id` (String) Synthetic ID (`system-disk` or `system-disk:<env>`).
- `response_json` (String) Full response payload as JSON.
- `disk_usage_json` (String) `diskUsage` object as JSON (or `null`).

