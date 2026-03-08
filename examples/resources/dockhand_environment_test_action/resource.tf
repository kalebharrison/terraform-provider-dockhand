resource "dockhand_environment_test_action" "direct" {
  connection_type = "direct"
  host            = "dind"
  port            = 2375
  protocol        = "http"
  fail_on_error   = true
  trigger         = "run-1"
}
