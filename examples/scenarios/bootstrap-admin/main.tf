terraform {
  required_providers {
    dockhand = {
      source  = "kalebharrison/dockhand"
      version = ">= 0.1.63"
    }
  }
}

provider "dockhand" {
  endpoint              = var.dockhand_endpoint
  allow_unauthenticated = true
}

resource "dockhand_user" "initial_admin" {
  username     = var.admin_username
  password     = var.admin_password
  email        = var.admin_email
  display_name = var.admin_display_name
  is_admin     = true
  is_active    = true
}
