# dockhand_system_file_content (Data Source)

Reads file content via `GET /api/system/files/content?path=<file>`.

Dockhand refuses protected host paths (`/etc`, `/proc`, `/root`, Dockhand DB/key files, and any `.ssh` / `.git` segment). Use a path under stack/data directories or another non-protected location.

## Example Usage

```hcl
data "dockhand_system_file_content" "compose" {
  path = "/docker/stacks/myapp/compose.yaml"
}
```

## Schema

### Required

- `path` (String) File path to read.

### Read-Only

- `id` (String) Synthetic ID (`system-file-content:<path>`).
- `content` (String) File content returned by Dockhand.
- `size` (Number) File size in bytes (or `null`).
- `mtime` (String) Last modification timestamp (or `null`).
- `result_json` (String) Full response payload as JSON.
