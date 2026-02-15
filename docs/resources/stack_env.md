# dockhand_stack_env (Resource)

Manages stack `.env` raw content and secret variables.

## Example Usage

```terraform
resource "dockhand_stack_env" "example" {
  env        = "2"
  stack_name = "my-stack"

  raw_content = <<-EOT
APP_ENV=prod
LOG_LEVEL=info
EOT

  secret_variables = [
    {
      key       = "API_TOKEN"
      value     = "super-secret"
      is_secret = true
    }
  ]
}
```

## Schema

### Required

- `stack_name` (String)

### Optional

- `env` (String)
- `raw_content` (String)
- `secret_variables` (List of Object)

### Read-Only

- `id` (String)
