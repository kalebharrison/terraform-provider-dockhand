resource "dockhand_environment" "example" {
  name            = "example-direct"
  connection_type = "direct"
  host            = "docker.internal"
  port            = 2375
  protocol        = "http"
  tls_skip_verify = false
  icon            = "globe"
  public_ip       = "203.0.113.10"

  collect_activity                 = true
  collect_metrics                  = true
  highlight_changes                = true
  update_check_enabled             = false
  update_check_auto_update         = false
  image_prune_enabled              = false
  vulnerability_scanning_enabled   = false
  ensure_grype_installed           = false
  ensure_trivy_installed           = false
  timezone                         = "UTC"
}
