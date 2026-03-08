# dockhand_system (Data Source)

Reads Dockhand system summary via `GET /api/system`.

## Example Usage

```hcl
data "dockhand_system" "current" {}
```

## Schema

### Read-Only

- `id` (String) Synthetic ID (`system`).
- `system_json` (String) Full response payload as JSON.
- `runtime_json` (String) Runtime subsection as JSON (or `null`).
- `database_json` (String) Database subsection as JSON (or `null`).
- `stats_json` (String) Stats subsection as JSON (or `null`).
- `docker_json` (String) Docker subsection as JSON (or `null`).
- `host_json` (String) Host subsection as JSON (or `null`).

