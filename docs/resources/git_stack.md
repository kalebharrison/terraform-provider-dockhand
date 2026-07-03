# Resource: dockhand_git_stack

Manages a Git-backed stack deployment in Dockhand via `/api/git/stacks`.

Use this resource when you want Dockhand to deploy/manage a stack from a Git repository in a target environment.

## Example Usage

```terraform
resource "dockhand_git_stack" "ollama" {
  env            = "11"
  stack_name     = "jetson01-ollama"
  repository_id  = "1"
  compose_path   = "stacks/jetson01/enabled/ollama/stack.yaml"
  deploy_now     = true
  build_on_deploy = true
  repull_images   = false
  force_redeploy  = false
}
```

Shared env files outside the compose directory:

```terraform
resource "dockhand_git_stack" "metube" {
  env             = "11"
  stack_name      = "metube"
  repository_id   = "1"
  compose_path    = "stacks/metube/compose.yml"
  context_dir     = "."
  env_file_path   = "./shared.env"
}
```

Set `context_dir` to widen Dockhand's compose working directory (relative to the repo root). This matches the Dockhand UI "Context directory" setting and is required when `env_file_path` or compose volume paths reference files outside the compose file folder. The compose file path must remain inside the context directory.

## Schema

### Required

- `stack_name` (String)
- `compose_path` (String)

### Optional

- `env` (String)
- `repository_id` (String)
- `repo_name` (String)
- `url` (String)
- `branch` (String, default: `main`)
- `credential_id` (String)
- `context_dir` (String) Repository-relative working directory for Docker Compose. Use `.` for the repo root when shared files live outside the compose file directory.
- `env_file_path` (String)
- `auto_update_enabled` (Boolean, default: `false`)
- `auto_update_cron` (String, default: `0 3 * * *`)
- `webhook_enabled` (Boolean, default: `false`)
- `webhook_secret` (String, Sensitive)
- `deploy_now` (Boolean, default: `false`) One-shot: set `true` to deploy on apply; state resets to `false` after apply.
- `build_on_deploy` (Boolean, default: `false`)
- `repull_images` (Boolean, default: `false`)
- `force_redeploy` (Boolean, default: `false`) One-shot: set `true` to force a redeploy on apply; state resets to `false` after apply.
- `env_vars_json` (String, Sensitive, default: `[]`) Write-only JSON array of env vars (`key`, `value`, `isSecret`). Dockhand does not echo values back; Terraform keeps the configured value in state.

`repository_id` is preferred when you already manage the repository with `dockhand_git_repository`.
If `repository_id` is not set, `url` must be set (and optional `repo_name`, `branch`, `credential_id`).

### Read-Only

- `id` (String)
- `last_sync` (String)
- `last_commit` (String)
- `sync_status` (String)
- `sync_error` (String)
- `created_at` (String)
- `updated_at` (String)
- `repository_name` (String)
- `repository_url` (String)
- `repository_branch` (String)

## Import

```bash
terraform import dockhand_git_stack.example <id>
```

Git stack ID. Set provider `default_env` or resource `env` before import.
