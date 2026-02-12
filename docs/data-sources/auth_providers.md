# dockhand_auth_providers (Data Source)

Reads configured authentication providers from `/api/auth/providers`.

## Example Usage

```terraform
data "dockhand_auth_providers" "this" {}
```

## Schema

### Read-Only

- `id` (String) Static ID: `auth-providers`.
- `default_provider` (String) Default provider ID.
- `providers` (List of Object)
  - `id` (String)
  - `name` (String)
  - `type` (String)

