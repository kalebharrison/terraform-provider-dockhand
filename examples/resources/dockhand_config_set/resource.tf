resource "dockhand_config_set" "example" {
  name        = "example-defaults"
  description = "Example container defaults"

  env_vars = {
    TZ = "UTC"
  }

  labels = {
    "com.example.owner" = "terraform"
  }

  ports = [
    {
      container_port = 80
      host_port      = 8080
      protocol       = "tcp"
    }
  ]

  network_mode   = "bridge"
  restart_policy = "no"
}
