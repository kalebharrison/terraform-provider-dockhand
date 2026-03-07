resource "dockhand_git_credential" "example" {
  name      = "github-token"
  auth_type = "password"
  username  = "x-access-token"
  password  = "replace-with-token"
}
