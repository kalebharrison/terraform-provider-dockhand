# dockhand_image

Manages a Dockhand image using pull/delete endpoints under `/api/images`.

## Example Usage

```hcl
resource "dockhand_image" "nginx" {
  name            = "nginx:latest"
  env             = "1"
  scan_after_pull = false
}
```

## Behavior

- `create` pulls the image using `/api/images/pull`.
- `read` resolves the image from `/api/images` (by ID, then tag match).
- `delete` removes the image using `/api/images/{id}`.

## Schema

### Required

- `name` (String) Image reference to pull.

### Optional

- `env` (String) Optional environment ID. If omitted, provider `default_env` is used.
- `scan_after_pull` (Boolean) Trigger scan during pull.

### Read-Only

- `id` (String) Image ID.
- `tags` (List of String) Tags currently reported by Dockhand.
- `size` (Number) Image size in bytes.
- `created_at` (String) Image creation timestamp (RFC3339).

## Import

```bash
terraform import dockhand_image.nginx <image-id>
```
