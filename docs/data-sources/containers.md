# dockhand_containers (Data Source)

Lists containers from Dockhand.

## Example Usage

```terraform
data "dockhand_containers" "all" {
  env = "2"
}
```

## Schema

### Optional

- `env` (String) Optional environment ID query parameter.

### Read-Only

- `containers` (List of Object)
  - `id` (String)
  - `name` (String)
  - `image` (String)
  - `state` (String)
  - `status` (String)
  - `health` (String)
  - `restart_count` (Number)
  - `command` (String)
  - `labels` (Map of String)
- `ids` (List of String) Sorted container IDs.
- `names` (List of String) Sorted container names.
