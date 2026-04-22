# Dockhand Provider

Use the Dockhand provider to manage Dockhand itself as code: bootstrap auth, configure Dockhand settings, register environments, wire Git and registries, and manage stacks, containers, images, volumes, networks, schedules, and operational actions.

## Quick Start

```terraform
terraform {
  required_providers {
    dockhand = {
      source  = "kalebharrison/dockhand"
      version = ">= 0.1.63"
    }
  }
}

provider "dockhand" {
  endpoint    = var.dockhand_endpoint
  username    = var.dockhand_username
  password    = var.dockhand_password
  default_env = var.dockhand_default_env
}
```

If you are using a Dockhand 1.0.25+ API token, use `api_token` instead of username/password:

```terraform
provider "dockhand" {
  endpoint  = var.dockhand_endpoint
  api_token = var.dockhand_api_token
}
```

For a fresh, unauthenticated Dockhand install, bootstrap the first admin with:

```terraform
provider "dockhand" {
  endpoint              = var.dockhand_endpoint
  allow_unauthenticated = true
}

resource "dockhand_user" "admin" {
  username     = "admin"
  password     = var.initial_admin_password
  email        = "admin@example.com"
  display_name = "Dockhand Admin"
  is_admin     = true
  is_active    = true
}
```

## Recommended Starting Points

- Bootstrap first admin: `examples/scenarios/bootstrap-admin/main.tf`
- Register an environment and deploy a stack: `examples/scenarios/environment-and-stack/main.tf`
- Configure Git-backed Dockhand automation: `examples/scenarios/gitops-stack/main.tf`
- Configure registries and image actions: `examples/scenarios/registry-and-image/main.tf`

## Coverage Highlights

- Dockhand settings: `dockhand_settings_general`, `dockhand_auth_settings`, `dockhand_notification`, `dockhand_license`, `dockhand_config_set`
- Environments: `dockhand_environment`, `dockhand_environment_test_action`, `dockhand_environment_scanner_action`
- Git: `dockhand_git_credential`, `dockhand_git_repository`, `dockhand_git_repository_test_action`, `dockhand_git_preview_env`, `dockhand_git_stack`, `dockhand_git_stack_webhook_action`, `dockhand_git_stack_deploy_action`, `dockhand_git_stack_env_file`
- Runtime resources: `dockhand_stack`, `dockhand_container`, `dockhand_image`, `dockhand_network`, `dockhand_volume`, `dockhand_schedule`
- Operational actions: `dockhand_batch_action`, `dockhand_prune_action`, stack/container/image/network/volume/schedule action resources
- Observability data sources: health, activity, users, environments, registries, containers, schedules, system, stack metadata, and jobs

## Compatibility and Known Gaps

- Compatibility policy: `docs/testing/compatibility-matrix.md`
- Release gate: `docs/testing/release-gate.md`
- Non-present endpoints and WebUI gaps: `docs/non-present-endpoints.md`
- API surface matrix: `docs/api-matrix.md`

## Project Standards

- Security policy: `SECURITY.md`
- Governance: `GOVERNANCE.md`
- Maintainers: `MAINTAINERS.md`
- Contributor support: `SUPPORT.md`
- Maintenance playbook: `docs/MAINTENANCE_PLAYBOOK.md`

## Resources

- `dockhand_stack`
- `dockhand_stack_action`
- `dockhand_user`
- `dockhand_settings_general`
- `dockhand_auth_settings`
- `dockhand_license`
- `dockhand_registry`
- `dockhand_registry_image_delete_action`
- `dockhand_git_credential`
- `dockhand_git_repository`
- `dockhand_git_repository_test_action`
- `dockhand_git_stack`
- `dockhand_git_stack_webhook_action`
- `dockhand_git_stack_deploy_action`
- `dockhand_git_stack_env_file`
- `dockhand_config_set`
- `dockhand_notification`
- `dockhand_notification_test_action`
- `dockhand_environment`
- `dockhand_environment_test_action`
- `dockhand_environment_scanner_action`
- `dockhand_network`
- `dockhand_network_connection_action`
- `dockhand_volume`
- `dockhand_volume_clone_action`
- `dockhand_image`
- `dockhand_image_push_action`
- `dockhand_image_scan_action`
- `dockhand_container`
- `dockhand_container_file`
- `dockhand_container_action`
- `dockhand_container_rename_action`
- `dockhand_container_update_action`
- `dockhand_container_check_updates_action`
- `dockhand_schedule`
- `dockhand_schedule_settings`
- `dockhand_schedule_run_action`
- `dockhand_prune_action`
- `dockhand_batch_action`
- `dockhand_stack_scan_action`
- `dockhand_stack_adopt_action`
- `dockhand_stack_env`

## Data Sources

- `dockhand_health`
- `dockhand_activity`
- `dockhand_hawser_status`
- `dockhand_users`
- `dockhand_registries`
- `dockhand_registry_search`
- `dockhand_registry_tags`
- `dockhand_registry_catalog`
- `dockhand_git_credentials`
- `dockhand_git_preview_env`
- `dockhand_git_repositories`
- `dockhand_notifications`
- `dockhand_config_sets`
- `dockhand_environments`
- `dockhand_environment_detect_socket`
- `dockhand_networks`
- `dockhand_volumes`
- `dockhand_images`
- `dockhand_auth_providers`
- `dockhand_schedules`
- `dockhand_schedule_settings`
- `dockhand_schedule_stream`
- `dockhand_schedules_executions`
- `dockhand_system`
- `dockhand_system_disk`
- `dockhand_system_files`
- `dockhand_system_file_content`
- `dockhand_containers`
- `dockhand_container_stats`
- `dockhand_container_pending_updates`
- `dockhand_container_shells`
- `dockhand_container_logs`
- `dockhand_container_inspect`
- `dockhand_container_processes`
- `dockhand_stack_sources`
- `dockhand_stack_base_path`
- `dockhand_stack_default_path`
- `dockhand_stacks`
- `dockhand_job`

## Schema

### Optional

- `endpoint` (String) Dockhand API base URL. Can also be set with `DOCKHAND_ENDPOINT`.
- `username` (String) Username for login-based auth. Can also be set with `DOCKHAND_USERNAME`.
- `password` (String, Sensitive) Password for login-based auth. Can also be set with `DOCKHAND_PASSWORD`.
- `api_token` (String, Sensitive) Dockhand API token for bearer-token auth. Can also be set with `DOCKHAND_API_TOKEN`.
- `mfa_token` (String, Sensitive) Optional MFA token for login-based auth. Can also be set with `DOCKHAND_MFA_TOKEN`.
- `auth_provider` (String) Auth provider ID (default `local`). Can also be set with `DOCKHAND_AUTH_PROVIDER`.
- `default_env` (String) Default environment ID used when resources omit `env`. Can also be set with `DOCKHAND_DEFAULT_ENV`.
- `insecure` (Boolean) Disable TLS verification.
- `allow_unauthenticated` (Boolean) Allow provider initialization without login credentials for first-install bootstrap flows. Can also be set with `DOCKHAND_ALLOW_UNAUTHENTICATED`.
