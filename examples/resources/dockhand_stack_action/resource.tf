resource "dockhand_stack_action" "restart_example" {
  env        = "2"
  stack_name = "replace-with-stack-name"
  action     = "restart"
  trigger    = "2026-02-12T03:00:00Z"
}
