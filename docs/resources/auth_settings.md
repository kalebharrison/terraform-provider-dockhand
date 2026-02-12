# dockhand_auth_settings

Manages Dockhand authentication settings via `/api/auth/settings`.

## Example Usage

```terraform
resource "dockhand_auth_settings" "this" {
  auth_enabled     = true
  default_provider = "local"
  session_timeout  = 86400
}
```

## Notes

- This is a singleton resource. The `id` is always `auth`.
- `delete` is a no-op because Dockhand does not expose a delete/reset endpoint for auth settings.

