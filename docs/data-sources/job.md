# dockhand_job (Data Source)

Reads an asynchronous Dockhand job via `/api/jobs/{jobId}`.

## Example Usage

```hcl
data "dockhand_job" "batch" {
  job_id = dockhand_batch_action.restart_containers.job_id
}
```

## Schema

### Required

- `job_id` (String) Dockhand async job ID.

### Read-Only

- `id` (String) Synthetic ID set to `job_id`.
- `status` (String) Current job status.
- `lines_json` (String) JSON array of job output lines.
- `result_json` (String) JSON object of job result payload.
