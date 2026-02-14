resource "dockhand_volume_clone_action" "example" {
  source_name = "source-volume"
  target_name = "source-volume-clone"
  trigger     = "manual-1"
}
