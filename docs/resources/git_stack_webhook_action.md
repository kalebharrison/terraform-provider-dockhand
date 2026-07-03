# dockhand_git_stack_webhook_action (Resource)

Triggers a one-shot webhook run for a Dockhand git stack.

## Example Usage

```terraform
resource "dockhand_git_stack_webhook_action" "sync_stack" {
  stack_id = "12"
  trigger  = "2026-02-12T22:00:00Z"
}
```

## Schema

### Required

- `stack_id` (String) Git stack ID.

### Optional

- `trigger` (String) Arbitrary value; change it to re-run.

### Read-Only

- `id` (String) Internal action execution ID.
## Import

Action resources are one-shot. Import the computed `id` (usually the Dockhand job ID or action record ID) after the action has run:

```bash
terraform import dockhand_git_stack_webhook_action.example <id>
```
