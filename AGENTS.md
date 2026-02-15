# AGENTS.md

## Purpose

Repository guidance for coding agents working on `terraform-provider-dockhand`.

## Project Context

- This repo builds a Terraform provider for Dockhand using the Terraform Plugin Framework.
- Current implemented surface:
  - Provider: `dockhand`
  - Resource: `dockhand_stack`
  - Resource: `dockhand_stack_action`
  - Resource: `dockhand_user`
  - Resource: `dockhand_notification`
  - Resource: `dockhand_settings_general`
  - Resource: `dockhand_auth_settings`
  - Resource: `dockhand_license`
  - Resource: `dockhand_registry`
  - Resource: `dockhand_git_credential`
  - Resource: `dockhand_git_repository`
  - Resource: `dockhand_git_stack_webhook_action`
  - Resource: `dockhand_git_stack_deploy_action`
  - Resource: `dockhand_config_set`
  - Resource: `dockhand_environment`
  - Resource: `dockhand_network`
  - Resource: `dockhand_network_connection_action`
  - Resource: `dockhand_volume`
  - Resource: `dockhand_volume_clone_action`
  - Resource: `dockhand_image`
  - Resource: `dockhand_image_push_action`
  - Resource: `dockhand_image_scan_action`
  - Resource: `dockhand_container`
  - Resource: `dockhand_container_file`
  - Resource: `dockhand_container_action`
  - Resource: `dockhand_container_rename_action`
  - Resource: `dockhand_container_update_action`
  - Resource: `dockhand_container_check_updates_action`
  - Resource: `dockhand_schedule`
  - Resource: `dockhand_stack_adopt_action`
  - Data source: `dockhand_health`
  - Data source: `dockhand_activity`
  - Data source: `dockhand_hawser_status`
  - Data source: `dockhand_stacks`
  - Data source: `dockhand_stack_sources`
  - Data source: `dockhand_containers`
  - Data source: `dockhand_container_stats`
  - Data source: `dockhand_container_pending_updates`
  - Data source: `dockhand_container_shells`
  - Resource: `dockhand_stack_scan_action`
  - Data source: `dockhand_container_logs`
  - Data source: `dockhand_container_inspect`
  - Data source: `dockhand_container_processes`
  - Data source: `dockhand_auth_providers`
  - Data source: `dockhand_schedules`
  - Data source: `dockhand_schedules_executions`

## Working Rules

- Keep changes focused and incremental.
- Prefer updating existing files over introducing new abstractions too early.
- Preserve backward compatibility for provider schema where practical.
- Do not commit secrets, tokens, or local override files.

## Code Standards

- Language: Go (module in `go.mod`).
- Keep code `gofmt` clean.
- Favor explicit error handling and actionable diagnostic messages.
- Keep provider/resource/data source schema docs aligned with behavior.

## Validation Commands

Run from repo root:

```bash
go mod tidy
go test ./...
go build ./...
```

If tests require Dockhand access, clearly separate them as acceptance tests.

## Release-First Workflow (Local Testing)

When making provider changes, prefer testing the exact same artifact that users will consume:

- Create a Git tag `vX.Y.Z` and push it.
- Wait for the `Release Artifacts` workflow to publish a GitHub Release and zip assets.
- Download the release zip(s) into the repo-local filesystem mirror at `./terraform/dockhand/mirror`.
- Run Terraform from `./terraform/dockhand/test` using `TF_CLI_CONFIG_FILE=../terraformrc.dockhand`.

Avoid testing by building local zips for release validation.

## Terraform Provider Notes

- Provider address is `registry.terraform.io/kalebharrison/dockhand` (see `main.go`). For private local development, use a Terraform CLI `dev_overrides` block and run Terraform via `scripts/tf-dev.sh` (skipping `terraform init`).
- Local dev workflow doc: `docs/LOCAL_DEV.md`.
- Provider config supports:
  - `endpoint`
  - `username`
  - `password` (sensitive)
  - `mfa_token` (optional sensitive)
  - `auth_provider` (optional, default `local`)
  - `default_env`
  - `insecure`
- Environment variable fallbacks:
  - `DOCKHAND_ENDPOINT`
  - `DOCKHAND_USERNAME`
  - `DOCKHAND_PASSWORD`
  - `DOCKHAND_MFA_TOKEN`
  - `DOCKHAND_AUTH_PROVIDER`
  - `DOCKHAND_DEFAULT_ENV`

## API Integration Notes

- Current client assumes:
  - `POST /api/auth/login`
  - `GET /api/auth/session`
  - `GET /api/stacks`
  - `POST /api/stacks`
  - `POST /api/stacks/{name}/start`
  - `POST /api/stacks/{name}/stop`
  - `DELETE /api/stacks/{name}?force=true`
  - `GET /api/dashboard/stats`
  - `GET/POST/DELETE /api/networks`
  - `POST /api/networks/{id}/connect`
  - `POST /api/networks/{id}/disconnect`
  - `GET/POST/DELETE /api/volumes`
  - `POST /api/volumes/{name}/clone`
  - `GET/POST/DELETE /api/images`
  - `POST /api/images/push`
  - `POST /api/images/scan`
  - `GET/POST/DELETE /api/containers`
  - `POST /api/containers/{id}/start`
  - `POST /api/containers/{id}/stop`
  - `POST /api/containers/{id}/restart`
  - `POST /api/containers/{id}/pause`
  - `POST /api/containers/{id}/unpause`
  - `POST /api/containers/{id}/rename`
  - `POST /api/containers/{id}/update`
  - `GET /api/containers/stats`
  - `POST /api/containers/check-updates`
  - `GET /api/containers/pending-updates`
  - `GET /api/containers/{id}/shells`
  - `GET /api/containers/{id}/logs`
  - `GET /api/containers/{id}/top`
  - `POST /api/containers/{id}/files/create`
  - `GET/PUT /api/containers/{id}/files/content`
  - `DELETE /api/containers/{id}/files/delete`
  - `GET /api/activity`
  - `GET /api/hawser/connect`
  - `POST /api/git/stacks/{id}/webhook`
  - `POST /api/git/stacks/{id}/deploy-stream`
  - `GET /api/schedules`
  - `GET /api/schedules/executions`
  - `POST /api/stacks/scan`
  - `POST /api/stacks/adopt`
  - `GET /api/stacks/sources`
  - `POST /api/schedules/system/{id}/toggle`
  - `POST /api/schedules/{type}/{id}/toggle`
- Verify response payload shapes against live Dockhand responses before release.

## CI Expectations

- GitHub Actions workflow: `.github/workflows/go-ci.yml`
- CI should pass:
  - format checks
  - `go mod tidy` consistency
  - tests
  - build

## Suggested Next Milestones

1. Add acceptance tests using `terraform-plugin-testing`.
2. Add release automation (private registry or public Terraform Registry artifacts).
3. Expand resource/data source coverage once API contracts are finalized.
