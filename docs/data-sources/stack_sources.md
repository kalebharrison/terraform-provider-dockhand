# dockhand_stack_sources (Data Source)

Reads stack source mappings from `/api/stacks/sources`.

## Example Usage

```terraform
data "dockhand_stack_sources" "this" {}
```

## Schema

### Read-Only

- `id` (String)
- `sources` (List of Object)
  - `stack_name` (String)
  - `source_type` (String)
  - `compose_path` (String)
  - `repository_id` (String)
  - `repository_name` (String)
  - `repository_url` (String)
  - `repository_branch` (String)
  - `repository_compose_path` (String)
  - `repository_environment_id` (String)
  - `repository_sync_status` (String)
