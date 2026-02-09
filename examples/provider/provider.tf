terraform {
  required_providers {
    dockhand = {
      # When developing locally, use scripts/tf-dev.sh (dev_overrides) rather than terraform init.
      source = "kalebharrison/dockhand"
    }
  }
}

provider "dockhand" {
  endpoint    = var.dockhand_endpoint
  username    = var.dockhand_username
  password    = var.dockhand_password
  default_env = var.dockhand_default_env
}
