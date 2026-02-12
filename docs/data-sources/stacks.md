# dockhand_stacks (Data Source)

Lists stacks from Dockhand.

## Example Usage

```terraform
data "dockhand_stacks" "all" {
  env = "2"
}
```

## Schema

### Optional

- `env` (String) Optional environment ID query parameter.

### Read-Only

- `stacks` (List of Object)
  - `name` (String)
  - `status` (String)
  - `container_count` (Number)
