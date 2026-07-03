# dockhand_stack_scan_action (Resource)

Runs a one-shot stack discovery scan.

## Example Usage

```terraform
resource "dockhand_stack_scan_action" "scan" {
  trigger = "2026-02-14T12:00:00Z"
}
```

## Schema

### Optional

- `trigger` (String) Arbitrary value; change it to re-run the scan.

### Read-Only

- `id` (String) Internal action execution ID.
- `discovered_count` (Number)
- `adopted_count` (Number)
- `skipped_count` (Number)
- `error_count` (Number)
- `result_json` (String) Raw JSON response payload.
## Import

Action resources are one-shot. Import the computed `id` (usually the Dockhand job ID or action record ID) after the action has run:

```bash
terraform import dockhand_stack_scan_action.example <id>
```
