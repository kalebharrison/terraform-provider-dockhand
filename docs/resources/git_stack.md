# Resource: dockhand_git_stack

Manages a Git-backed stack deployment in Dockhand via `/api/git/stacks`.

Use this resource when you want Dockhand to deploy/manage a stack from a Git repository in a target environment.

## Example Usage

```terraform
resource "dockhand_git_stack" "ollama" {
  env           = "11"
  stack_name    = "jetson01-ollama"
  repository_id = "1"
  compose_path  = "stacks/jetson01/enabled/ollama/stack.yaml"
  deploy_now    = true
}
```

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
- `env_file_path` (String)
- `auto_update_enabled` (Boolean, default: `false`)
- `auto_update_cron` (String, default: `0 3 * * *`)
- `webhook_enabled` (Boolean, default: `false`)
- `webhook_secret` (String, Sensitive)
- `deploy_now` (Boolean, default: `false`)
- `env_vars_json` (String, default: `[]`)

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
