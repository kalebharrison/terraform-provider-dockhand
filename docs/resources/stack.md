# dockhand_stack (Resource)

Manages a Dockhand stack.

## Example Usage

```terraform
resource "dockhand_stack" "example" {
  name = "hello"
  env  = "1"
  compose = <<-YAML
    services:
      nginx:
        image: nginx:stable
        ports:
          - "8080:80"
  YAML
  enabled = true
}
```

## Schema

### Required

- `name` (String) Stack name.
- `compose` (String) Stack compose manifest content.

### Optional

- `env` (String) Optional environment ID query parameter.
- `enabled` (Boolean) Whether the stack should be running. Defaults to `true`.

### Read-Only

- `id` (String) Stack ID (`<env>:<name>` or `<name>`).

## Import

Import by stack ID:

```bash
terraform import dockhand_stack.example <name>

# or with explicit env
terraform import dockhand_stack.example <env>:<name>
```
