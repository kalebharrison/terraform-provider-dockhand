# Terraform Provider Dockhand

Terraform provider for managing Dockhand resources.

## Current Scope

This initial scaffold includes:

- Provider config with `endpoint`, `username`/`password` (login-based), `default_env`, and `insecure`.
- Resource: `dockhand_stack`
- Resource: `dockhand_user`
- Resource: `dockhand_settings_general`
- Resource: `dockhand_registry`
- Resource: `dockhand_git_credential`
- Resource: `dockhand_git_repository`
- Resource: `dockhand_config_set`
- Data source: `dockhand_health`
- HTTP client wiring against:
  - `POST /api/auth/login` (session-based auth)
  - `GET /api/auth/session` (session check)
  - `GET /api/stacks`
  - `POST /api/stacks`
  - `POST /api/stacks/{name}/start`
  - `POST /api/stacks/{name}/stop`
  - `DELETE /api/stacks/{name}?force=true`
  - `GET /api/dashboard/stats` (health signal)

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
go test -v ./internal/provider -run TestAccUserResource
```

## Release

This repo currently focuses on private/local development and private distribution.

To publish versioned zip artifacts to a GitHub Release (useful for downloading and then installing into a local/team mirror), push a tag like `v0.1.0`.

See `docs/PRIVATE_DISTRIBUTION.md` for installing from a filesystem mirror.
