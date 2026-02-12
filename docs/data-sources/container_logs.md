# dockhand_container_logs (Data Source)

Fetches container logs from Dockhand.

## Example Usage

```terraform
data "dockhand_container_logs" "web" {
  env          = "2"
  container_id = "abc123..."
  tail         = 200
}
```

## Schema

### Required

- `container_id` (String) Container ID to read logs for.

### Optional

- `env` (String) Optional environment ID query parameter.
- `tail` (Number) Number of log lines to request. Defaults to `100`.

### Read-Only

- `logs` (String) Returned log content.
