resource "dockhand_schedule_run_action" "run_now" {
  type        = "system_cleanup"
  schedule_id = "2"
  trigger     = "example-run-1"
}
