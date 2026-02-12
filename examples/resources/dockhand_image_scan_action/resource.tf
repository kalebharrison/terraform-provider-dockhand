resource "dockhand_image_scan_action" "scan_example" {
  env        = "2"
  image_name = "redis:7-alpine"
  trigger    = "2026-02-12T03:00:00Z"
}
