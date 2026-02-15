# Terraform Provider Dockhand

Terraform provider for managing Dockhand resources.

## Current Scope

This initial scaffold includes:

- Provider config with `endpoint`, `username`/`password` (login-based), `default_env`, and `insecure`.
- Resource: `dockhand_stack`
- Resource: `dockhand_stack_action`
- Resource: `dockhand_user`
- Resource: `dockhand_settings_general`
- Resource: `dockhand_registry`
- Resource: `dockhand_git_credential`
- Resource: `dockhand_git_repository`
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
- Resource: `dockhand_schedule_run_action`
- Resource: `dockhand_stack_env`
- Data source: `dockhand_health`
- Data source: `dockhand_activity`
- Data source: `dockhand_hawser_status`
- Data source: `dockhand_environments`
- Data source: `dockhand_networks`
- Data source: `dockhand_volumes`
- Data source: `dockhand_images`
- Data source: `dockhand_stacks`
- Data source: `dockhand_container_logs`
- Data source: `dockhand_container_inspect`
- Data source: `dockhand_container_processes`
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
  - `POST /api/git/stacks/{id}/webhook`
  - `GET /api/git/stacks/{id}/env-files`
  - `POST /api/git/stacks/{id}/env-files`
  - `GET /api/stacks/{name}/env`
  - `PUT /api/stacks/{name}/env`
  - `GET /api/stacks/{name}/env/raw`
  - `PUT /api/stacks/{name}/env/raw`
  - `GET /api/schedules`
  - `POST /api/schedules/{type}/{id}/run`
  - `POST /api/schedules/system/{id}/toggle`
  - `POST /api/schedules/{type}/{id}/toggle`

If your Dockhand API differs, update `internal/provider/client.go`.

## Development

Requirements:

- Go 1.22+
- Terraform CLI

Build:

```bash
go mod tidy
go test ./...
go build ./...
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
go test -v ./internal/provider -run 'TestAcc(ContainerFileDirectoryResourceTerraform|ContainerProcessesDataSourceTerraform|StackActionDownTerraform|StackEnvResourceTerraform|GitStackEnvFileResourceTerraform)'
```

## Release

This repo currently focuses on private/local development and private distribution.

To publish versioned zip artifacts to a GitHub Release (useful for downloading and then installing into a local/team mirror), push a tag like `v0.1.0`.

See `docs/PRIVATE_DISTRIBUTION.md` for installing from a filesystem mirror.
