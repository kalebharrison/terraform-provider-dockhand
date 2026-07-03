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
## Import

Action resources are one-shot. Import the computed `id` (usually the Dockhand job ID or action record ID) after the action has run:

```bash
terraform import dockhand_git_stack_deploy_action.example <id>
```
