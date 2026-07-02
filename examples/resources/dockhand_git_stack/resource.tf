resource "dockhand_git_stack" "example" {
  env             = "2"
  stack_name      = "example-git-stack"
  repository_id   = "1"
  compose_path    = "stacks/example/stack.yaml"
  context_dir     = "."
  env_file_path   = "./shared.env"
  deploy_now      = true
  build_on_deploy = true
  repull_images   = false
  force_redeploy  = false
}
