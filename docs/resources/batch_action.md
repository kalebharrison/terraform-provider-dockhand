# dockhand_batch_action (Resource)

Runs an asynchronous batch operation in Dockhand via `/api/batch` and optionally waits for completion.

## Example Usage

```hcl
resource "dockhand_batch_action" "restart_containers" {
  env                 = "2"
  entity_type         = "containers"
  operation           = "restart"
  item_ids            = ["container-id-1", "container-id-2"]
  wait_for_completion = true
  timeout_seconds     = 60
  poll_interval_ms    = 1000
  trigger             = "manual-run-1"
}
```

## Schema

### Required

- `entity_type` (String) Batch entity type (for example, `containers`).
- `item_ids` (List of String) IDs to include in the batch operation.
- `operation` (String) Batch operation name (for example, `restart`).

### Optional

- `env` (String) Optional environment ID query parameter.
- `poll_interval_ms` (Number) Poll interval in milliseconds when waiting for completion.
- `timeout_seconds` (Number) Maximum wait time in seconds when waiting for completion.
- `trigger` (String) Change this value to run the action again.
- `wait_for_completion` (Boolean) Whether to wait for terminal job completion. Defaults to `true`.

### Read-Only

- `id` (String) Synthetic ID set to the Dockhand job ID.
- `job_id` (String) Dockhand async job ID returned by `/api/batch`.
- `job_status` (String) Current or terminal job status from `/api/jobs/{jobId}`.
- `lines_json` (String) JSON array of job output lines.
- `result_json` (String) JSON object of job result payload.
