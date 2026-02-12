# dockhand_container_inspect (Data Source)

Reads the full container inspect payload from Dockhand.

## Example Usage

```terraform
data "dockhand_container_inspect" "web" {
  env          = "2"
  container_id = "abc123..."
}
```

## Schema

### Required

- `container_id` (String) Container ID to inspect.

### Optional

- `env` (String) Optional environment ID query parameter.

### Read-Only

- `inspect_json` (String) Full JSON inspect payload.
