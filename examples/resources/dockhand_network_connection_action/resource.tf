resource "dockhand_network_connection_action" "example" {
  network_id   = "2f4e9f3d20f6"
  container_id = "a1b2c3d4e5f6"
  action       = "connect"
  trigger      = "manual-1"
}
