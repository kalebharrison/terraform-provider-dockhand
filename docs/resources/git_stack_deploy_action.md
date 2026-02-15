# dockhand_git_stack_deploy_action (Resource)

Runs a one-shot Git stack deploy request via `/api/git/stacks/{id}/deploy-stream`.

## Example Usage

```terraform
resource "dockhand_git_stack_deploy_action" "deploy" {
  stack_id = 12
  trigger  = timestamp()
}
```

## Schema

### Required

- `stack_id` (String)

### Optional

- `trigger` (String)

### Read-Only

- `id` (String)
- `result` (String)
- `output` (String)
