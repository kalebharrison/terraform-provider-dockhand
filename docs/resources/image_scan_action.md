# dockhand_image_scan_action (Resource)

Runs a one-shot vulnerability scan for an image.

## Example Usage

```terraform
resource "dockhand_image_scan_action" "scan_redis" {
  env        = "2"
  image_name = "redis:7-alpine"
  trigger    = "2026-02-12T03:00:00Z"
}
```

## Schema

### Required

- `image_name` (String) Image tag/name to scan.

### Optional

- `env` (String) Optional environment ID query parameter.
- `trigger` (String) Arbitrary value; change it to re-run the scan.

### Read-Only

- `id` (String) Internal action execution ID.
- `result` (String) Scan request result marker.
