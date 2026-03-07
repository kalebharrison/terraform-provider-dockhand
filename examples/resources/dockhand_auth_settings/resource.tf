resource "dockhand_auth_settings" "example" {
  auth_enabled     = true
  default_provider = "local"
  session_timeout  = 86400
}
