# dockhand_volume_clone_action (Resource)

Runs a one-shot volume clone action.

## Example Usage

```terraform
resource "dockhand_volume_clone_action" "clone" {
  source_name = "source-volume"
  target_name = "source-volume-clone"
  trigger     = "2026-02-14T10:00:00Z"
}
```

## Schema

### Required

- `source_name` (String) Existing source volume name.
- `target_name` (String) New cloned volume name.

### Optional

- `env` (String) Optional environment ID query parameter.
- `trigger` (String) Arbitrary value; change it to re-run the action.

### Read-Only

- `id` (String) Internal action execution ID.
