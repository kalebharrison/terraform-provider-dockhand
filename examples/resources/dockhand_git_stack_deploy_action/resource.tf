resource "dockhand_git_stack_deploy_action" "example" {
  stack_id = "12"
  trigger  = timestamp()
}
