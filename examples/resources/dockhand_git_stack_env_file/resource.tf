resource "dockhand_git_stack_env_file" "example" {
  stack_id = "12"
  path     = "stacks/app/.env"
  trigger  = timestamp()
}
