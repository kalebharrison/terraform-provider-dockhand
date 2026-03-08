# dockhand_environment_detect_socket (Data Source)

Reads Dockhand socket discovery output via `GET /api/environments/detect-socket`.

## Example Usage

```hcl
data "dockhand_environment_detect_socket" "local" {}
```

## Schema

### Read-Only

- `id` (String) Synthetic ID (`dockhand-environment-detect-socket`).
- `home_dir` (String) Dockhand container home directory from probe response.
- `socket_paths` (List of String) Best-effort extracted socket paths from `sockets` payload.
- `sockets_json` (String) Raw `sockets` payload as JSON.
