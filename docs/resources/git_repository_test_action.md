# Resource: dockhand_git_repository_test_action

Runs a one-shot Git repository connectivity test via `POST /api/git/repositories/test`.

Change `trigger` to force Terraform to run the test again.

## Example Usage

```terraform
resource "dockhand_git_repository_test_action" "example" {
  url           = "https://github.com/docker/awesome-compose.git"
  branch        = "master"
  fail_on_error = true
  trigger       = "run-1"
}
```

## Schema

### Optional

- `repository_id` (String) Existing Dockhand Git repository ID to test. Mutually exclusive with `url`.
- `url` (String) Git clone URL to test when `repository_id` is not used.
- `branch` (String) Git branch to test when `url` is used.
- `credential_id` (String) Optional Dockhand Git credential ID to test with `url`.
- `compose_path` (String) Optional compose path included in the test payload when `url` is used.
- `fail_on_error` (Boolean) Default `true`. If true, apply fails when Dockhand returns `success = false`.
- `trigger` (String) Force rerun marker.

### Read-Only

- `id` (String) Internal action instance ID.
- `success` (Boolean) Dockhand repository test success flag.
- `error` (String) Dockhand error message when present.
- `resolved_branch` (String) Branch Dockhand actually tested.
- `last_commit` (String) Last commit hash returned by Dockhand when the test succeeds.
- `result_json` (String) Raw normalized result JSON from the test response.
