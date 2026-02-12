# dockhand_stack_action (Resource)

Runs a one-shot action on a stack.

## Example Usage

```terraform
resource "dockhand_stack_action" "restart_stack" {
  env        = "2"
  stack_name = "nextcloud"
  action     = "restart"
  trigger    = "2026-02-12T03:00:00Z"
}
```

## Schema

### Required

- `stack_name` (String) Stack name to act on.
- `action` (String) Action to execute: `start`, `stop`, or `restart`.

### Optional

- `env` (String) Optional environment ID query parameter.
- `trigger` (String) Arbitrary value; change it to re-run the action.

### Read-Only

- `id` (String) Internal action execution ID.
