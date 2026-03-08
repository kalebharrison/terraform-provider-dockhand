# dockhand_registry_catalog (Data Source)

Reads catalog payloads from `GET /api/registry/catalog`.

## Example Usage

```hcl
data "dockhand_registry_catalog" "dockerhub" {
  registry = "1"
  page     = 1
  page_size = 100
}
```

## Schema

### Optional

- `registry` (String) Optional registry selector.
- `page` (Number) Optional page number.
- `page_size` (Number) Optional page size.

### Read-Only

- `id` (String) Synthetic data source ID (`<registry>:<page>:<page_size>`).
- `repository_count` (Number) Number of extracted repository names.
- `repositories` (List of String) Best-effort repository names extracted from catalog payload.
- `catalog_json` (String) Raw catalog payload as JSON.
