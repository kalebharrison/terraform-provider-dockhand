resource "dockhand_git_stack_webhook_action" "example" {
  stack_id       = "12"
  webhook_secret = "replace-with-stack-webhook-secret"
  trigger        = "2026-02-12T22:00:00Z"
}
