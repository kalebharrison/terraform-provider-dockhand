# Resource: dockhand_notification_test_action

Runs a one-shot notification test via `POST /api/notifications/test`.

Change `trigger` to force Terraform to run the test again.

## Example Usage

```terraform
resource "dockhand_notification_test_action" "smtp" {
  type = "smtp"
  config_json = jsonencode({
    host       = "smtp.example.local"
    port       = 25
    from_email = "dockhand@example.local"
    to_emails  = ["ops@example.local"]
  })

  fail_on_error = false
  trigger       = "run-1"
}
```

## Schema

### Optional

- `notification_id` (String) Existing Dockhand notification ID to test.
- `type` (String) Notification type when `notification_id` is not set.
- `config_json` (String, Sensitive) JSON object for test `config` when `notification_id` is not set.
- `fail_on_error` (Boolean) Default `true`. If true, apply fails when test returns `success = false`.
- `trigger` (String) Force rerun marker.

### Read-Only

- `id` (String) Internal action instance ID (`test:<source>:<trigger>`).
- `success` (Boolean) Notification test success flag.
- `error` (String) Error message returned by Dockhand.
- `message` (String) Informational message returned by Dockhand.
- `result_json` (String) Raw test result payload as JSON.
## Import

Action resources are one-shot. Import the computed `id` (usually the Dockhand job ID or action record ID) after the action has run:

```bash
terraform import dockhand_notification_test_action.example <id>
```
