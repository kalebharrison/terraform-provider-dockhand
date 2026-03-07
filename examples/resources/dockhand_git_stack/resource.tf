resource "dockhand_git_stack" "example" {
  env           = "2"
  stack_name    = "example-git-stack"
  repository_id = "1"
  compose_path  = "stacks/example/stack.yaml"
  deploy_now    = true
}
