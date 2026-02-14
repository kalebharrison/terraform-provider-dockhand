# dockhand_image_push_action (Resource)

Runs a one-shot image push to a Dockhand registry.

## Example Usage

```terraform
resource "dockhand_image_push_action" "push" {
  image_id    = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  registry_id = 1
  trigger     = "2026-02-14T10:00:00Z"
}
```

## Schema

### Required

- `image_id` (String) Image ID to push.
- `registry_id` (Number) Dockhand registry ID.

### Optional

- `env` (String) Optional environment ID query parameter.
- `trigger` (String) Arbitrary value; change it to re-run the action.

### Read-Only

- `id` (String) Internal action execution ID.
- `result` (String) Push request result marker.
