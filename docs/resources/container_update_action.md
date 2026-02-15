# dockhand_container_update_action (Resource)

Runs a one-shot container update with a raw JSON payload.

## Example Usage

```terraform
resource "dockhand_container_update_action" "update" {
  env          = "2"
  container_id = "abc123..."
  payload_json = jsonencode({})
  trigger      = "manual-1"
}
```

## Schema

### Required

- `container_id` (String)

### Optional

- `env` (String)
- `payload_json` (String) JSON object sent to `/api/containers/{id}/update`.
- `restart_policy_name` (String) Typed helper for `RestartPolicy.Name`.
- `restart_policy_maximum_retry_count` (Number) Typed helper for `RestartPolicy.MaximumRetryCount`.
- `cpu_shares` (Number) Typed helper for `CpuShares`.
- `pids_limit` (Number) Typed helper for `PidsLimit`.
- `memory_bytes` (Number) Typed helper for `Memory`.
- `nano_cpus` (Number) Typed helper for `NanoCpus`.
- `trigger` (String)

### Read-Only

- `id` (String)
- `result_json` (String)
