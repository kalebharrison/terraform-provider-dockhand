# dockhand_schedule_stream (Data Source)

Reads a bounded snapshot from Dockhand schedules stream (`GET /api/schedules/stream`).

This data source does not keep a live stream open in Terraform state. It captures up to `max_events` within `timeout_seconds` and returns those events.

## Example Usage

```hcl
data "dockhand_schedule_stream" "snapshot" {
  max_events      = 1
  timeout_seconds = 2
}
```

## Schema

### Optional

- `max_events` (Number) Max events to capture. Default `1`.
- `timeout_seconds` (Number) Max read duration. Default `5`.

### Read-Only

- `id` (String) Synthetic ID (`schedule-stream:<max_events>:<timeout_seconds>`).
- `connected` (Boolean) Whether a `connected` stream event was observed.
- `event_count` (Number) Number of captured events.
- `events_json` (String) Raw captured events as JSON array.
- `schedules_json` (String) Raw schedules payload from captured `schedules` event (if present).
- `events` (List of Object) Captured SSE events:
  - `event` (String)
  - `data` (String)
