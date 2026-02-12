resource "dockhand_container" "example" {
  name  = "tf-nginx"
  env   = "2"
  image = "nginx:alpine"

  enabled        = true
  network_mode   = "bridge"
  restart_policy = "unless-stopped"

  env_vars = {
    NGINX_ENTRYPOINT_QUIET_LOGS = "1"
  }

  labels = {
    managed_by = "terraform"
  }

  ports = [
    {
      container_port = 80
      host_port      = "18080"
      protocol       = "tcp"
    }
  ]
}
