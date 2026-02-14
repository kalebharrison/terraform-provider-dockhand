# dockhand_container_shells (Data Source)

Reads available terminal shells for a container from Dockhand.

## Example Usage

```terraform
data "dockhand_container_shells" "web" {
  env          = "2"
  container_id = "abc123..."
}
```

## Schema

### Required

- `container_id` (String) Container ID.

### Optional

- `env` (String) Optional environment ID.

### Read-Only

- `shells` (List of String) Available shell paths.
- `default_shell` (String) Dockhand-selected default shell (if present).
- `all_shells` (List of Object)
  - `path` (String)
  - `label` (String)
  - `available` (Boolean)
