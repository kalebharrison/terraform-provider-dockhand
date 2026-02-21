# Resource: dockhand_environment_scanner_action

Runs one-shot scanner operations for a Dockhand environment.

Supported actions:

- `install_grype`
- `install_trivy`
- `remove_grype`
- `remove_trivy`
- `check_updates`

Change `trigger` to force Terraform to run the action again.

## Example Usage

```terraform
resource "dockhand_environment_scanner_action" "install_grype" {
  env     = "2"
  action  = "install_grype"
  trigger = "run-1"
}

resource "dockhand_environment_scanner_action" "check_updates" {
  env     = "2"
  action  = "check_updates"
  trigger = "run-1"
}
```

## Schema

### Required

- `action` (String) One of `install_grype`, `install_trivy`, `remove_grype`, `remove_trivy`, `check_updates`.

### Optional

- `env` (String) Environment ID. Falls back to provider `default_env` when omitted.
- `trigger` (String) Force rerun marker.

### Read-Only

- `id` (String) Internal action instance ID (`<env>:<action>:<trigger>`).
- `grype_installed` (Boolean) Scanner availability state after action.
- `trivy_installed` (Boolean) Scanner availability state after action.
- `grype_version` (String) Scanner version reported by Dockhand.
- `trivy_version` (String) Scanner version reported by Dockhand.
- `grype_has_update` (Boolean) Only populated by `check_updates` action.
- `trivy_has_update` (Boolean) Only populated by `check_updates` action.
- `result_json` (String) Raw action summary payload from provider execution.
