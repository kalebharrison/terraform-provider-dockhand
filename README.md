# Terraform Provider Dockhand

Terraform provider for managing Dockhand resources.

## Current Scope

This provider currently includes:

- Provider config with `endpoint`, `username`/`password` (login-based), `default_env`, `insecure`, and `allow_unauthenticated` (bootstrap mode).
- Resource: `dockhand_stack`
- Resource: `dockhand_stack_action`
- Resource: `dockhand_user`
- Resource: `dockhand_settings_general`
- Resource: `dockhand_registry`
- Resource: `dockhand_registry_image_delete_action`
- Resource: `dockhand_git_credential`
- Resource: `dockhand_git_repository`
- Resource: `dockhand_git_stack`
- Resource: `dockhand_git_stack_webhook_action`
- Resource: `dockhand_git_stack_deploy_action`
- Resource: `dockhand_git_stack_env_file`
- Resource: `dockhand_config_set`
- Resource: `dockhand_network`
- Resource: `dockhand_volume`
- Resource: `dockhand_image`
- Resource: `dockhand_image_scan_action`
- Resource: `dockhand_container`
- Resource: `dockhand_container_file`
- Resource: `dockhand_container_action`
- Resource: `dockhand_schedule`
- Resource: `dockhand_schedule_settings`
- Resource: `dockhand_schedule_run_action`
- Resource: `dockhand_prune_action`
- Resource: `dockhand_batch_action`
- Resource: `dockhand_stack_env`
- Resource: `dockhand_stack_scan_action`
- Resource: `dockhand_stack_adopt_action`
- Resource: `dockhand_auth_settings`
- Resource: `dockhand_license`
- Resource: `dockhand_notification`
- Resource: `dockhand_notification_test_action`
- Resource: `dockhand_environment`
- Resource: `dockhand_environment_test_action`
- Resource: `dockhand_environment_scanner_action`
- Resource: `dockhand_network_connection_action`
- Resource: `dockhand_volume_clone_action`
- Resource: `dockhand_image_push_action`
- Resource: `dockhand_container_rename_action`
- Resource: `dockhand_container_update_action`
- Resource: `dockhand_container_check_updates_action`
- Data source: `dockhand_health`
- Data source: `dockhand_activity`
- Data source: `dockhand_hawser_status`
- Data source: `dockhand_auth_providers`
- Data source: `dockhand_schedules`
- Data source: `dockhand_schedule_settings`
- Data source: `dockhand_schedule_stream`
- Data source: `dockhand_schedules_executions`
- Data source: `dockhand_containers`
- Data source: `dockhand_container_stats`
- Data source: `dockhand_container_pending_updates`
- Data source: `dockhand_container_shells`
- Data source: `dockhand_stack_sources`
- Data source: `dockhand_stack_base_path`
- Data source: `dockhand_stack_default_path`
- Data source: `dockhand_container_logs`
- Data source: `dockhand_container_inspect`
- Data source: `dockhand_container_processes`
- Data source: `dockhand_stacks`
- Data source: `dockhand_users`
- Data source: `dockhand_registries`
- Data source: `dockhand_registry_search`
- Data source: `dockhand_registry_tags`
- Data source: `dockhand_registry_catalog`
- Data source: `dockhand_git_credentials`
- Data source: `dockhand_git_repositories`
- Data source: `dockhand_notifications`
- Data source: `dockhand_config_sets`
- Data source: `dockhand_environments`
- Data source: `dockhand_environment_detect_socket`
- Data source: `dockhand_networks`
- Data source: `dockhand_volumes`
- Data source: `dockhand_images`
- Data source: `dockhand_job`
- HTTP client wiring against:
  - `POST /api/auth/login` (session-based auth)
  - `GET /api/auth/session` (session check)
  - `GET /api/stacks`
  - `POST /api/stacks`
  - `POST /api/stacks/{name}/start`
  - `POST /api/stacks/{name}/stop`
  - `DELETE /api/stacks/{name}?force=true`
  - `GET /api/containers/{id}/logs`
  - `GET /api/dashboard/stats` (health signal)
  - `GET/POST/DELETE /api/networks`
  - `GET/POST/DELETE /api/volumes`
  - `GET/POST/DELETE /api/images`
  - `POST /api/images/scan`
  - `GET/POST/DELETE /api/containers`
  - `POST /api/containers/{id}/start`
  - `POST /api/containers/{id}/stop`
  - `POST /api/containers/{id}/restart`
  - `GET /api/activity`
  - `GET /api/hawser/connect`
  - `POST /api/notifications/test`
  - `GET /api/registry/search`
  - `GET /api/registry/tags`
  - `GET /api/registry/catalog`
  - `DELETE /api/registry/image`
  - `GET /api/environments/detect-socket`
  - `POST /api/environments/test`
  - `POST /api/git/stacks/{id}/webhook`
  - `GET /api/git/stacks/{id}/env-files`
  - `POST /api/git/stacks/{id}/env-files`
  - `GET /api/stacks/{name}/env`
  - `PUT /api/stacks/{name}/env`
  - `GET /api/stacks/{name}/env/raw`
  - `PUT /api/stacks/{name}/env/raw`
  - `GET /api/stacks/base-path`
  - `GET /api/stacks/default-path`
  - `GET /api/schedules`
  - `GET /api/schedules/settings`
  - `PUT /api/schedules/settings`
  - `GET /api/schedules/stream`
  - `POST /api/schedules/{type}/{id}/run`
  - `POST /api/schedules/system/{id}/toggle`
  - `POST /api/schedules/{type}/{id}/toggle`
  - `POST /api/prune/all`
  - `POST /api/prune/containers`
  - `POST /api/prune/images`
  - `POST /api/prune/networks`
  - `POST /api/prune/volumes`
  - `POST /api/batch`
  - `GET /api/jobs/{id}`

If your Dockhand API differs, update `internal/provider/client.go`.

## Development

Requirements:

- Go 1.22+
- Terraform CLI

Build:

```bash
./scripts/verify.sh
# or:
make verify
```

Run locally with Terraform:

```hcl
terraform {
  required_providers {
    dockhand = {
      # Address this provider serves as today (see main.go ServeOpts.Address).
      # This does not require publishing when using Terraform dev_overrides.
      source = "kalebharrison/dockhand"
    }
  }
}

provider "dockhand" {
  endpoint       = "https://dockhand.example.com"
  username       = var.dockhand_username
  password       = var.dockhand_password
  default_env    = "1"
}

# Fresh-install bootstrap mode (auth disabled): allow unauthenticated calls so the first admin can be created.
provider "dockhand" {
  endpoint              = "https://dockhand.example.com"
  allow_unauthenticated = true
}
```

Local development (private, no registry publish):

```bash
REPO="/path/to/terraform-provider-dockhand"
(cd "$REPO" && make tf-dev-build)

# In your Terraform config directory:
export DOCKHAND_ENDPOINT="http://dockhand.example.internal:13001"
export DOCKHAND_USERNAME="your-username"
export DOCKHAND_PASSWORD="your-password"
export DOCKHAND_DEFAULT_ENV="1"

"$REPO/scripts/tf-dev.sh" plan
"$REPO/scripts/tf-dev.sh" apply
```

Private distribution (team-friendly, still private):

- Filesystem mirror workflow: `docs/PRIVATE_DISTRIBUTION.md`
- Endpoint contract probe workflow: `docs/ENDPOINT_PROBE.md`
- Public/registry readiness checklist: `docs/REGISTRY_READINESS.md`
- Compatibility matrix and recursive validation: `docs/testing/compatibility-matrix.md`
- Release gate policy: `docs/testing/release-gate.md`
- Maintainer runbook (standard change/release loop): `docs/MAINTENANCE_PLAYBOOK.md`

Repository policy and governance:

- `SECURITY.md`
- `CODE_OF_CONDUCT.md`
- `SUPPORT.md`
- `GOVERNANCE.md`
- `MAINTAINERS.md`

Example resource:

```hcl
resource "dockhand_stack" "example" {
  name = "nextcloud"
  env  = "1"
  compose = <<-YAML
    services:
      app:
        image: nextcloud:latest
  YAML
  enabled = true
}
```

User resource example:

```hcl
resource "dockhand_user" "example" {
  username     = "tf-example-user"
  password     = var.dockhand_user_password
  email        = "tf-example-user@example.local"
  display_name = "Terraform Example User"
  is_admin     = false
  is_active    = true
}
```

Bootstrap (fresh install with auth disabled):

```hcl
provider "dockhand" {
  endpoint              = "http://dockhand.example.internal:13001"
  allow_unauthenticated = true
}

resource "dockhand_user" "initial_admin" {
  username     = "admin"
  password     = var.initial_admin_password
  email        = "admin@example.local"
  display_name = "Initial Admin"
  is_admin     = true
  is_active    = true
}
```

After the initial admin exists, switch back to normal provider auth (`username`/`password`) for ongoing Terraform management.

## Acceptance Tests

User acceptance tests are environment-gated and require real Dockhand access:

```bash
export DOCKHAND_TEST_ENDPOINT="http://dockhand.example.internal:13001"
export DOCKHAND_TEST_USERNAME="your-username"
export DOCKHAND_TEST_PASSWORD="your-password"
export DOCKHAND_TEST_DEFAULT_ENV="1"
go test -v ./internal/provider -run 'TestAcc(UserResource|ContainerRenameAction)'

# Optional container update action acceptance test (uses an existing container fixture):
export DOCKHAND_TEST_UPDATE_CONTAINER_ID="existing-container-id"
go test -v ./internal/provider -run 'TestAccContainerUpdateAction'

# New surfaces acceptance tests (optional env-gated cases):
# - Requires an existing running container for directory tests:
export DOCKHAND_TEST_FILE_CONTAINER_ID="existing-container-id"
# - Requires a git-managed stack id and an env-file path in that stack repo:
export DOCKHAND_TEST_GIT_STACK_ID="12"
export DOCKHAND_TEST_GIT_STACK_ENV_PATH="stacks/app/.env"
# - Requires a resolvable host name for Docker-in-Docker direct environment tests:
export DOCKHAND_TEST_DIND_HOST="dind"
# - Requires an active Hawser edge agent token for agent-environment connectivity tests:
export DOCKHAND_TEST_AGENT_TOKEN="ci-agent-token"
go test -v ./internal/provider -run 'TestAcc(ContainerFileDirectoryResourceTerraform|ContainerProcessesDataSourceTerraform|StackActionDownTerraform|StackEnvResourceTerraform|GitStackEnvFileResourceTerraform|BatchActionAndJobDataSourceTerraform|BatchActionNoWaitTerraform)'

# Direct environment connectivity acceptance test (Dockhand -> DinD):
go test -v ./internal/provider -run 'TestAccEnvironmentResourceDirectDinDTerraform'

# Agent-token + Hawser connectivity acceptance test:
go test -v ./internal/provider -run 'TestAccEnvironmentResourceAgentTokenTerraform'
```

GitHub Actions acceptance workflow:

- Workflow: `.github/workflows/acceptance-ci.yml`
- Spins up:
  - `fnsys/dockhand:latest`
  - `docker:27-dind`
- Bootstraps first admin credentials for the ephemeral Dockhand instance.
- Creates a dedicated direct environment (`ci-dind`) targeting the DinD container.
- Launches a Hawser edge agent inside DinD with a generated token.
- Runs targeted acceptance tests, including environment connectivity coverage.

Full recursive acceptance harness:

- Local script: `scripts/run-acceptance-harness.sh`
- Nightly workflow: `.github/workflows/acceptance-full.yml`
- Dockhand release watcher: `.github/workflows/dockhand-release-watch.yml`

## Release

This repo currently focuses on private/local development and private distribution.

To publish versioned zip artifacts to a GitHub Release (useful for downloading and then installing into a local/team mirror), push a tag like `v0.1.0`.

For Terraform Registry readiness, configure GitHub repository secrets used by the release workflow:

- `GPG_PRIVATE_KEY` (ASCII-armored private key)
- `GPG_PASSPHRASE` (private key passphrase)

See `docs/PRIVATE_DISTRIBUTION.md` for installing from a filesystem mirror.
