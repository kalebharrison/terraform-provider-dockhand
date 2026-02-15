# dockhand_git_stack_env_file (Resource)

Reads env vars for a selected git stack env file.

## Example Usage

```terraform
resource "dockhand_git_stack_env_file" "example" {
  stack_id = "12"
  path     = "stacks/app/.env"
  trigger  = timestamp()
}
```

## Schema

### Required

- `stack_id` (String)
- `path` (String)

### Optional

- `trigger` (String)

### Read-Only

- `id` (String)
- `vars_json` (String)
- `file_paths` (List of String)
