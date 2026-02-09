# dockhand_user (Resource)

Manages a Dockhand user via `/api/users`.

## Example Usage

```terraform
resource "dockhand_user" "example" {
  username     = "tf-user"
  password     = var.user_password
  email        = "tf-user@example.local"
  display_name = "Terraform User"
  is_admin     = false
  is_active    = true
}
```

## Schema

### Required

- `username` (String) Username.

### Optional

- `password` (String, Sensitive) Password (write-only).
- `email` (String) Email.
- `display_name` (String) Display name.
- `is_admin` (Boolean) Admin flag. Defaults to `false`.
- `is_active` (Boolean) Active flag. Defaults to `true`.

### Read-Only

- `id` (String) User ID.
- `mfa_enabled` (Boolean) Whether MFA is enabled.
- `last_login` (String) Last login timestamp.
- `created_at` (String) Created timestamp.
- `updated_at` (String) Updated timestamp.

## Import

Import by user ID:

```bash
terraform import dockhand_user.example <user-id>
```
