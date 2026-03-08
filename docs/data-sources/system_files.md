# dockhand_system_files (Data Source)

Lists files/directories for a path via `GET /api/system/files?path=<dir>`.

## Example Usage

```hcl
data "dockhand_system_files" "root" {
  path = "/"
}
```

## Schema

### Optional

- `path` (String) Directory path to list. Defaults to `/` when omitted.

### Read-Only

- `id` (String) Synthetic ID (`system-files:<path>`).
- `parent` (String) Parent directory path (or `null` for root/no parent).
- `entry_count` (Number) Number of entries returned.
- `entries_json` (String) Array of entry objects as JSON.
- `result_json` (String) Full response payload as JSON.

