resource "dockhand_environment_scanner_action" "install_grype" {
  env     = "2"
  action  = "install_grype"
  trigger = "example-run-1"
}
