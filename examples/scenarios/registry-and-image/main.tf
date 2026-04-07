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

resource "dockhand_registry" "ghcr" {
  name     = "ghcr"
  provider = "docker-v2"
  url      = "https://ghcr.io"
  username = var.registry_username
  password = var.registry_password
}

resource "dockhand_image" "app" {
  env        = var.dockhand_default_env
  name       = var.image_name
  pull       = true
  registry_id = dockhand_registry.ghcr.id
}

resource "dockhand_image_scan_action" "app" {
  env      = var.dockhand_default_env
  image_id = dockhand_image.app.id
}
