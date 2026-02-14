resource "dockhand_image_push_action" "example" {
  image_id    = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  registry_id = 1
  trigger     = "manual-1"
}
