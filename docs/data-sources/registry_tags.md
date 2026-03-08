# dockhand_registry_tags (Data Source)

Lists image tags via `GET /api/registry/tags`.

## Example Usage

```hcl
data "dockhand_registry_tags" "busybox" {
  image    = "library/busybox"
  registry = "1"
  page     = 1
  page_size = 25
}
```

## Schema

### Required

- `image` (String) Repository image path (for example `library/busybox`).

### Optional

- `registry` (String) Optional registry selector.
- `page` (Number) Optional page number.
- `page_size` (Number) Optional page size.

### Read-Only

- `id` (String) Synthetic data source ID (`<image>:<registry>:<page>:<page_size>`).
- `total` (Number) Total tags reported by Dockhand.
- `has_next` (Boolean) Whether there is a next page.
- `has_prev` (Boolean) Whether there is a previous page.
- `tag_names` (List of String) Flattened tag name list.
- `tags_json` (String) Raw tag array as JSON.
- `tags` (List of Object) Parsed tag entries:
  - `name` (String)
  - `size` (Number)
  - `last_updated` (String)
  - `digest` (String)
