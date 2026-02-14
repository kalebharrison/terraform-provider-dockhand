# dockhand_container_stats (Data Source)

Reads live container stats from `/api/containers/stats`.

## Example Usage

```terraform
data "dockhand_container_stats" "this" {
  env = "2"
}
```

## Schema

### Optional

- `env` (String) Optional environment ID query parameter.

### Read-Only

- `id` (String)
- `stats` (List of Object)
  - `id` (String)
  - `name` (String)
  - `cpu_percent` (Number)
  - `memory_usage` (Number)
  - `memory_raw` (Number)
  - `memory_cache` (Number)
  - `memory_limit` (Number)
  - `memory_percent` (Number)
  - `network_rx` (Number)
  - `network_tx` (Number)
  - `block_read` (Number)
  - `block_write` (Number)
