resource "dockhand_git_repository_test_action" "example" {
  url           = "https://github.com/docker/awesome-compose.git"
  branch        = "master"
  fail_on_error = true
  trigger       = "run-1"
}
