resource "dockhand_registry_image_delete_action" "example" {
  registry      = "1"
  image         = "library/busybox"
  tag           = "latest"
  fail_on_error = false
  trigger       = "run-1"
}
