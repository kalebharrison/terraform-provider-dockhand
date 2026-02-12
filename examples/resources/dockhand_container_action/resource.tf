resource "dockhand_container_action" "restart_example" {
  env          = "2"
  container_id = "replace-with-container-id"
  action       = "restart"
  trigger      = "2026-02-12T03:00:00Z"
}
