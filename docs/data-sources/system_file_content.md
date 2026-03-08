# dockhand_system_file_content (Data Source)

Reads file content via `GET /api/system/files/content?path=<file>`.

## Example Usage

```hcl
data "dockhand_system_file_content" "hostname" {
  path = "/etc/hostname"
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

