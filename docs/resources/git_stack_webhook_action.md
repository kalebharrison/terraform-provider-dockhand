# dockhand_git_stack_webhook_action (Resource)

Triggers a one-shot webhook run for a Dockhand git stack.

## Example Usage

```terraform
resource "dockhand_git_stack_webhook_action" "sync_stack" {
  stack_id       = dockhand_git_stack.example.id
  webhook_secret = dockhand_git_stack.example.webhook_secret
  trigger        = "2026-02-12T22:00:00Z"
}
```

## Schema

### Required

- `stack_id` (String) Git stack ID.

### Optional

- `webhook_secret` (String, Sensitive) Webhook secret from the git stack. Required by current Dockhand when the stack webhook is enabled (request is signed with `X-Hub-Signature-256` / `X-Gitlab-Token` and `?secret=`).
- `trigger` (String) Arbitrary value; change it to re-run.

### Read-Only

- `id` (String) Internal action execution ID.
## Import

Action resources are one-shot. Import the computed `id` (usually the Dockhand job ID or action record ID) after the action has run:

```bash
terraform import dockhand_git_stack_webhook_action.example <id>
```
