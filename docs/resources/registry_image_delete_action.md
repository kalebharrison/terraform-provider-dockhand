# Resource: dockhand_registry_image_delete_action

Runs one-shot remote image tag deletion via `DELETE /api/registry/image`.

Change `trigger` to force Terraform to execute the delete action again.

## Example Usage

```hcl
resource "dockhand_registry_image_delete_action" "cleanup" {
  registry = "1"
  image    = "library/busybox"
  tag      = "latest"

  # Set false if your target registry does not allow deletes and you only
  # want observability instead of apply failure.
  fail_on_error = false
  trigger       = "run-1"
}
```

## Schema

### Required

- `registry` (String) Registry selector value expected by Dockhand.
- `image` (String) Repository image path.
- `tag` (String) Image tag to delete.

### Optional

- `fail_on_error` (Boolean) Default `true`. If true, apply fails on non-2xx responses.
- `trigger` (String) Change this value to run the action again.

### Read-Only

- `id` (String) Internal action instance ID (<registry>:<image>:<tag>:<trigger>).
- `success` (Boolean) True when Dockhand returned a success status.
- `status_code` (Number) HTTP status code returned by Dockhand.
- `error` (String) Error message when the request fails.
- `result_json` (String) Raw success/error payload as JSON.
## Import

Action resources are one-shot. Import the computed `id` (usually the Dockhand job ID or action record ID) after the action has run:

```bash
terraform import dockhand_registry_image_delete_action.example <id>
```
