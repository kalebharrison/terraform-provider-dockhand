resource "dockhand_schedule" "system_cleanup" {
  schedule_id = "2"
  type        = "system_cleanup"
  enabled     = true
}
