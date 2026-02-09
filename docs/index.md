# Dockhand Provider

Use the Dockhand provider to manage stack definitions and query API health.

## Example Usage

```terraform
provider "dockhand" {
  endpoint       = "https://dockhand.example.com"
  username       = var.dockhand_username
  password       = var.dockhand_password
  default_env    = "1"
}
```

## Resources

- `dockhand_stack`
- `dockhand_user`
- `dockhand_settings_general`
- `dockhand_registry`

## Schema

### Optional

- `endpoint` (String) Dockhand API base URL. Can also be set with `DOCKHAND_ENDPOINT`.
- `username` (String) Username for login-based auth. Can also be set with `DOCKHAND_USERNAME`.
- `password` (String, Sensitive) Password for login-based auth. Can also be set with `DOCKHAND_PASSWORD`.
- `mfa_token` (String, Sensitive) Optional MFA token for login-based auth. Can also be set with `DOCKHAND_MFA_TOKEN`.
- `auth_provider` (String) Auth provider id (default `local`). Can also be set with `DOCKHAND_AUTH_PROVIDER`.
- `default_env` (String) Default environment ID used when resources omit `env`. Can also be set with `DOCKHAND_DEFAULT_ENV`.
- `insecure` (Boolean) Disable TLS verification.
