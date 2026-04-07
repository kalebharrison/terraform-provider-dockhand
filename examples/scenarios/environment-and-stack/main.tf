terraform {
  required_providers {
    dockhand = {
      source  = "kalebharrison/dockhand"
      version = ">= 0.1.63"
    }
  }
}

provider "dockhand" {
  endpoint    = var.dockhand_endpoint
  username    = var.dockhand_username
  password    = var.dockhand_password
  default_env = var.dockhand_default_env
}

resource "dockhand_environment" "socket_env" {
  name            = var.environment_name
  connection_type = "socket"
  socket_path     = "/var/run/docker.sock"
  icon            = "server"
  timezone        = "UTC"
}

resource "dockhand_settings_general" "ui" {
  date_format                = "YYYY-MM-DD"
  time_format                = "24h"
  show_stopped_containers    = true
  highlight_available_updates = true
}

resource "dockhand_stack" "whoami" {
  env  = dockhand_environment.socket_env.id
  name = "whoami"
  compose = <<-YAML
    services:
      whoami:
        image: traefik/whoami:latest
        restart: unless-stopped
        ports:
          - "8080:80"
  YAML
  enabled = true
}

output "environment_id" {
  value = dockhand_environment.socket_env.id
}

output "stack_status" {
  value = dockhand_stack.whoami.status
}
