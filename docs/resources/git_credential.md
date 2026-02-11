# dockhand_git_credential

Manages a Dockhand Git credential via `/api/git/credentials`.

## Example Usage

Password/token credential:

```terraform
resource "dockhand_git_credential" "github" {
  name      = "github"
  auth_type = "password"

  # For GitHub tokens, the common pattern is:
  # username = "x-access-token"
  username = "x-access-token"
  password = var.github_token
}
```

If you are managing an existing credential and do not want to rotate the stored secret, omit `password` (Dockhand will keep the existing one).

## Schema

### Read-only

- `id` (String) Credential ID.
- `has_password` (Boolean)
- `has_ssh_key` (Boolean)
- `created_at` (String)
- `updated_at` (String)

### Required

- `name` (String)
- `auth_type` (String)

### Optional

- `username` (String)
- `ssh_key` (String, Sensitive, Write-only)
- `password` (String, Sensitive, Write-only)
