# dockhand_stack_default_path (Data Source)

Calculates default stack paths via `GET /api/stacks/default-path?name=<stack>`.

## Example Usage

```hcl
data "dockhand_stack_default_path" "example" {
  stack_name = "media-stack"
}
```

## Schema

### Required

- `stack_name` (String) Stack name used for path calculation.

### Read-Only

- `id` (String) Synthetic ID (`stack-default-path:<stack_name>`).
- `stack_dir` (String) Default stack directory.
- `compose_path` (String) Default compose file path.
- `env_path` (String) Default `.env` file path.
- `source` (String) Path resolution source indicator (if provided by Dockhand).
- `result_json` (String) Raw response payload as JSON.
