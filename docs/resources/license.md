# dockhand_license

Manages Dockhand licensing via `/api/license`.

## Example Usage

```terraform
resource "dockhand_license" "this" {
  name = var.dockhand_license_name
  key  = var.dockhand_license_key
}
```

## Notes

- This is a singleton resource. The `id` is always `license`.
- `name` and `key` are only required when you want Terraform to set/update the license.
- If you omit `name` and `key`, the resource reads and reports current license status only.
- `delete` revokes the current license via `DELETE /api/license`.

