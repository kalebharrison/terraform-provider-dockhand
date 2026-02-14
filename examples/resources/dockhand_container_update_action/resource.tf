resource "dockhand_container_update_action" "example" {
  env          = "2"
  container_id = "abc123..."
  payload_json = jsonencode({})
  trigger      = "manual-1"
}
