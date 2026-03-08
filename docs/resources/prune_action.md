# Resource: dockhand_prune_action

Runs one-shot cleanup actions via Dockhand prune endpoints:

- `POST /api/prune/all`
- `POST /api/prune/containers`
- `POST /api/prune/images`
- `POST /api/prune/networks`
- `POST /api/prune/volumes`

Change `trigger` to run prune again.

## Example Usage

```hcl
resource "dockhand_prune_action" "containers" {
  env                 = "1"
  mode                = "containers"
  wait_for_completion = false
  trigger             = "run-1"
}
```

## Schema

### Required

- `mode` (String) One of: `all`, `containers`, `images`, `networks`, `volumes`.

### Optional

- `env` (String) Optional environment ID query parameter.
- `wait_for_completion` (Boolean) Default `true`. Waits for terminal status when prune returns `jobId`.
- `timeout_seconds` (Number) Default `120`. Poll timeout for async jobs.
- `poll_interval_ms` (Number) Default `1000`. Poll interval for async jobs.
- `trigger` (String) Change this value to execute prune again.

### Read-Only

- `id` (String) Internal action instance ID.
- `success` (Boolean) True when prune completed without failure status.
- `status_code` (Number) HTTP status code returned by Dockhand.
- `job_id` (String) Async job ID when Dockhand returns one (for example image prune).
- `job_status` (String) Async job status (`submitted`, `running`, `done`, etc.) or `done`/`failed` for sync responses.
- `result_json` (String) Raw prune result payload JSON.
- `lines_json` (String) Async job output lines JSON array.
- `error` (String) Error message if a run fails.
