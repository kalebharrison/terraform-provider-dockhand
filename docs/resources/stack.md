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

  # Optional. The id comes from Dockhand's secret provider list.
  # secret_provider_path is the bulk selector (DOCKHAND_SECRET_SELECTOR).
  # secret_provider_id   = 3
  # secret_provider_path = "op://homelab/stack"
}
```

## Schema

### Required

- `name` (String) Stack name.
- `compose` (String) Stack compose manifest content.

### Optional

- `env` (String) Optional environment ID query parameter.
- `enabled` (Boolean) Whether the stack should be running. Defaults to `true`.
- `secret_provider_id` (Number) Dockhand secret provider id. Omit it to leave an existing binding unchanged. Remove a previously set value to clear the binding.
- `secret_provider_path` (String) Bulk selector written as `DOCKHAND_SECRET_SELECTOR`. This is the path field beside the provider dropdown. A legacy `OP_ENVIRONMENT_ID` row is replaced. Omit it to leave stack env vars unchanged.

### Read-Only

- `container_count` (Number) Current number of containers in the stack.
- `container_ids` (List of String) Container IDs currently associated with the stack.
- `id` (String) Stack ID (<env>:<name> or <name>).
- `status` (String) Current stack runtime status from Dockhand.

## Import

Import by stack ID:

```bash
terraform import dockhand_stack.example <name>

# or with explicit env
terraform import dockhand_stack.example <env>:<name>
```
