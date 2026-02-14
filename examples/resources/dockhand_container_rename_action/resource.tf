resource "dockhand_container_rename_action" "example" {
  env          = "2"
  container_id = "abc123..."
  name         = "new-container-name"
  trigger      = "manual-1"
}
