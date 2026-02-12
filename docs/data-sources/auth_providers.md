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

## Notes

- On free/local setups this typically returns `local` and any free providers (such as SSO/OIDC if enabled).
- LDAP/AD and role-management provider surfaces are currently treated as license-tier features and are not modeled as Terraform resources in this provider.
