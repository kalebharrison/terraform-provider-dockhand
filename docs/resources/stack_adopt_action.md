# dockhand_stack_adopt_action (Resource)

Runs a one-shot stack adopt operation.

## Example Usage

```terraform
resource "dockhand_stack_adopt_action" "adopt" {
  environment_id = 2
  trigger        = "manual-1"

  stacks = [
    {
      name         = "example"
      compose_path = "/app/data/stacks/example/compose.yaml"
    }
  ]
}
```

## Schema

### Required

- `environment_id` (Number)
- `stacks` (List of Object)
  - `name` (String)
  - `compose_path` (String)

### Optional

- `trigger` (String)

### Read-Only

- `id` (String)
- `adopted` (List of String)
- `failed` (List of String)
