# dockhand_stack_base_path (Data Source)

Reads global stack base path via `GET /api/stacks/base-path`.

## Example Usage

```hcl
data "dockhand_stack_base_path" "current" {}
```

## Schema

### Read-Only

- `id` (String) Synthetic ID (`stack-base-path`).
- `base_path` (String) Current Dockhand base path for stack directories.
