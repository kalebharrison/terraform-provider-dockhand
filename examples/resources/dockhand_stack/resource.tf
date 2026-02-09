resource "dockhand_stack" "nextcloud" {
  name = "nextcloud"
  env  = "1"
  compose = <<-YAML
    services:
      app:
        image: nextcloud:latest
        ports:
          - "8080:80"
  YAML
  enabled = true
}
