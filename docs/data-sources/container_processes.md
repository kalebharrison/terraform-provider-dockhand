# dockhand_container_processes (Data Source)

Reads process table information for a running container from `/api/containers/{id}/top`.

## Example Usage

```terraform
data "dockhand_container_processes" "top" {
  env          = "2"
  container_id = "abc123"
}
```

## Schema

### Required

- `container_id` (String)

### Optional

- `env` (String)

### Read-Only

- `id` (String)
- `titles` (List of String)
- `processes` (List of Object)
- `error` (String)
