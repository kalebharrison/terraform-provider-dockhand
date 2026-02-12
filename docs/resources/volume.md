# dockhand_volume

Manages a Dockhand volume via `/api/volumes`.

## Example Usage

```hcl
resource "dockhand_volume" "data" {
  name   = "tf-app-data"
  driver = "local"
  env    = "1"
  driver_options = {
    type = "none"
  }
  labels = {
    managed_by = "terraform"
  }
}
```

## Schema

### Required

- `name` (String) Volume name.

### Optional

- `driver` (String) Volume driver. Defaults to `local`.
- `env` (String) Optional environment ID. If omitted, provider `default_env` is used.
- `driver_options` (Map of String) Driver options map.
- `labels` (Map of String) Labels map.

### Read-Only

- `id` (String) Volume identifier (same as `name`).
- `mountpoint` (String) Host mountpoint.
- `scope` (String) Volume scope.
- `created_at` (String) Creation timestamp.

## Import

```bash
terraform import dockhand_volume.data <volume-name>
```
