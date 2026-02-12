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
- `dockhand_stack_action`
- `dockhand_user`
- `dockhand_settings_general`
- `dockhand_auth_settings`
- `dockhand_license`
- `dockhand_registry`
- `dockhand_git_credential`
- `dockhand_git_repository`
- `dockhand_git_stack_webhook_action`
- `dockhand_config_set`
- `dockhand_notification`
- `dockhand_environment`
- `dockhand_network`
- `dockhand_volume`
- `dockhand_image`
- `dockhand_image_scan_action`
- `dockhand_container`
- `dockhand_container_action`
- `dockhand_schedule`

## Data Sources

- `dockhand_health`
- `dockhand_activity`
- `dockhand_hawser_status`
- `dockhand_auth_providers`
- `dockhand_schedules`
- `dockhand_container_logs`
- `dockhand_container_inspect`
- `dockhand_stacks`

## Schema

### Optional

- `endpoint` (String) Dockhand API base URL. Can also be set with `DOCKHAND_ENDPOINT`.
- `username` (String) Username for login-based auth. Can also be set with `DOCKHAND_USERNAME`.
- `password` (String, Sensitive) Password for login-based auth. Can also be set with `DOCKHAND_PASSWORD`.
- `mfa_token` (String, Sensitive) Optional MFA token for login-based auth. Can also be set with `DOCKHAND_MFA_TOKEN`.
- `auth_provider` (String) Auth provider id (default `local`). Can also be set with `DOCKHAND_AUTH_PROVIDER`.
- `default_env` (String) Default environment ID used when resources omit `env`. Can also be set with `DOCKHAND_DEFAULT_ENV`.
- `insecure` (Boolean) Disable TLS verification.
