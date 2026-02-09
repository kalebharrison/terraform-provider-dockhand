resource "dockhand_user" "example" {
  username     = "tf-example-user"
  password     = var.dockhand_user_password
  email        = "tf-example-user@example.local"
  display_name = "Terraform Example User"
  is_admin     = false
  is_active    = true
}
