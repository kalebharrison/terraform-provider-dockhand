# dockhand_registry_search (Data Source)

Searches registry repositories via `GET /api/registry/search`.

## Example Usage

```hcl
data "dockhand_registry_search" "busybox" {
  term = "busybox"
}
```

## Schema

### Required

- `term` (String) Search term for the registry query.

### Optional

- `registry` (String) Optional registry selector (ID/name expected by Dockhand).

### Read-Only

- `id` (String) Synthetic data source ID (`<term>:<registry>`).
- `result_count` (Number) Number of search results returned by Dockhand.
- `image_names` (List of String) Best-effort extracted repository names from result objects.
- `results_json` (String) Raw search result array as JSON.
