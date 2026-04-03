data "dockhand_git_preview_env" "example" {
  url          = "https://github.com/docker/awesome-compose.git"
  branch       = "master"
  compose_path = "nginx-flask-mysql/compose.yaml"
}
