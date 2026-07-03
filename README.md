# Terraform Provider Dockhand

Manage Dockhand itself with Terraform: bootstrap the first admin, configure Dockhand settings, register environments, wire Git and registries, and manage stacks, containers, images, networks, volumes, schedules, and operational actions through one provider.

## Why This Provider

Dockhand already centralizes Docker operations. This provider lets you treat Dockhand as code as well:

- Bootstrap a fresh Dockhand install and create the first admin user.
- Manage Dockhand settings, authentication, notifications, license, and config sets.
- Register and test Docker environments, including socket, TCP/TLS, and Hawser agent connectivity.
- Manage registries, Git credentials, Git repositories, Git stacks, and stack environment files.
- Manage stacks, containers, images, networks, volumes, schedules, prune jobs, and batch actions.
- Query operational state with data sources for health, activity, containers, schedules, system state, registries, and more.

## Quick Start

```hcl
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

resource "dockhand_environment" "docker_socket" {
  name            = "truenas02"
  connection_type = "socket"
  socket_path     = "/var/run/docker.sock"
}

resource "dockhand_stack" "whoami" {
  env  = dockhand_environment.docker_socket.id
  name = "whoami"
  compose = <<-YAML
    services:
      whoami:
        image: traefik/whoami:latest
        ports:
          - "8080:80"
  YAML
  enabled = true
}

data "dockhand_health" "this" {}
```

## Authentication Modes

The provider supports Dockhand's login-based authentication flow and bearer-token API auth.

- Standard login: `endpoint`, `username`, `password`
- API token auth: `endpoint`, `api_token`
- Optional MFA: `mfa_token`
- Optional provider selection: `auth_provider`
- First-install bootstrap: `allow_unauthenticated = true`
- Computed endpoints: when `endpoint` references another resource's output, plan defers provider configuration until apply; login and API requests retry transient connection failures (configurable via `request_retry_*`)

Bootstrap example for a fresh Dockhand instance with no auth configured yet:

```hcl
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

## Coverage Summary

The provider currently covers these Dockhand areas:

| Area | Coverage |
| --- | --- |
| Dockhand core settings | `dockhand_settings_general`, `dockhand_auth_settings`, `dockhand_notification`, `dockhand_license`, `dockhand_config_set` |
| Environments | `dockhand_environment`, `dockhand_environment_test_action`, `dockhand_environment_scanner_action`, `dockhand_environment_detect_socket` |
| Git and Git stacks | `dockhand_git_credential`, `dockhand_git_repository`, `dockhand_git_repository_test_action`, `dockhand_git_preview_env`, `dockhand_git_stack`, `dockhand_git_stack_webhook_action`, `dockhand_git_stack_deploy_action`, `dockhand_git_stack_env_file` |
| Registries and images | `dockhand_registry`, `dockhand_registry_image_delete_action`, `dockhand_registry_search`, `dockhand_registry_tags`, `dockhand_registry_catalog`, `dockhand_image`, `dockhand_image_push_action`, `dockhand_image_scan_action` |
| Networks and volumes | `dockhand_network`, `dockhand_network_connection_action`, `dockhand_volume`, `dockhand_volume_clone_action` |
| Containers | `dockhand_container`, `dockhand_container_file`, `dockhand_container_action`, `dockhand_container_rename_action`, `dockhand_container_update_action`, `dockhand_container_check_updates_action` |
| Stacks | `dockhand_stack`, `dockhand_stack_action`, `dockhand_stack_env`, `dockhand_stack_scan_action`, `dockhand_stack_adopt_action` |
| Schedules and operations | `dockhand_schedule`, `dockhand_schedule_settings`, `dockhand_schedule_run_action`, `dockhand_prune_action`, `dockhand_batch_action`, `dockhand_job`, `dockhand_activity`, `dockhand_system*`, `dockhand_health` |

Full resource and data source coverage is documented in `docs/index.md` and `docs/api-matrix.md`.

## Scenario Examples

Task-oriented examples live under `examples/scenarios`:

- `examples/scenarios/bootstrap-admin` — first-install bootstrap of the initial admin user
- `examples/scenarios/environment-and-stack` — create an environment and deploy a stack into it
- `examples/scenarios/gitops-stack` — configure Git credentials, repository integration, Git stack, and env file management
- `examples/scenarios/registry-and-image` — configure a registry and manage an image lifecycle action flow

Per-resource examples remain under `examples/resources` and `examples/data-sources`.

## Known Gaps

This provider aims for broad Dockhand coverage, but not every WebUI surface maps cleanly to stable Terraform semantics yet.

Current known limitations and non-present endpoints are tracked in:

- `docs/non-present-endpoints.md`
- `docs/reports/webui-endpoint-gap-audit.md`
- `docs/testing/compatibility-matrix.md`

## Compatibility and Validation

The provider is continuously validated against real Dockhand instances.

- Recursive acceptance coverage runs through the Dockhand + Docker-in-Docker harness.
- A scheduled compatibility watcher checks new Dockhand releases and opens compatibility issues on drift.
- Endpoint, WebUI, docs-reference, and private-endpoint probes are tracked in repository reports.

Operational references:

- `docs/testing/compatibility-matrix.md`
- `docs/testing/release-gate.md`
- `docs/MAINTENANCE_PLAYBOOK.md`

## Development and CI

Requirements for optional local work:

- Go 1.25+
- Terraform CLI or OpenTofu

**Routine validation runs in GitHub Actions** (ephemeral Dockhand + DinD). See `docs/AGENT_AUTONOMY.md`.

Standard local gate (no Dockhand):

```bash
./scripts/verify.sh --quality
```

When dependency, toolchain, or security posture changes:

```bash
./scripts/verify.sh --security
```

Dockhand-dependent probes and acceptance run on CI. Local runs are debug-only.

Optional local Terraform/provider dev:

- `docs/LOCAL_DEV.md`
- `docs/PRIVATE_DISTRIBUTION.md`
- `docs/REGISTRY_READINESS.md`

## Provider Configuration Reference

Provider arguments:

- `endpoint` — Dockhand API base URL
- `username` — login username
- `password` — login password
- `api_token` — Dockhand API token for bearer-token auth
- `mfa_token` — optional MFA token
- `auth_provider` — optional auth provider ID, defaults to `local`
- `default_env` — default environment ID for resources that omit `env`
- `insecure` — disable TLS verification
- `allow_unauthenticated` — permit bootstrap flows before auth is configured
- `request_retry_attempts` — max attempts for login/API requests on transient connection failures (default `6`)
- `request_retry_min_seconds` — minimum wait between retries in seconds (default `1`)
- `request_retry_max_seconds` — maximum wait between retries in seconds (default `5`)

Environment variable fallbacks:

- `DOCKHAND_ENDPOINT`
- `DOCKHAND_USERNAME`
- `DOCKHAND_PASSWORD`
- `DOCKHAND_API_TOKEN`
- `DOCKHAND_MFA_TOKEN`
- `DOCKHAND_AUTH_PROVIDER`
- `DOCKHAND_DEFAULT_ENV`
- `DOCKHAND_ALLOW_UNAUTHENTICATED`
- `DOCKHAND_REQUEST_RETRY_ATTEMPTS`
- `DOCKHAND_REQUEST_RETRY_MIN_SECONDS`
- `DOCKHAND_REQUEST_RETRY_MAX_SECONDS`

## Additional Repository Policy

- `SECURITY.md`
- `GOVERNANCE.md`
- `MAINTAINERS.md`
- `SUPPORT.md`
- `CODE_OF_CONDUCT.md`

## Agent-managed workflow

Some changes land through an automated agent loop (Cursor Cloud Agent + GitHub Actions):

- Docs: `docs/AGENT_RUNBOOK.md`, `docs/AGENT_CODING_STANDARDS.md`, `docs/AGENT_IDENTITY.md`, `docs/AGENT_INTAKE.md`
- Branches: `agent/issue-<number>-<slug>`
- Commits include `Co-authored-by: Cursor Agent <noreply@cursor.com>`
- PRs are opened by GitHub Actions as `github-actions[bot]` with the `agent` label

One-time setup after enabling agent CI: `docs/AGENT_DEPLOYMENT.md`.
