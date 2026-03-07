resource "dockhand_git_credential" "example" {
  name      = "github-ssh"
  auth_type = "ssh"
  username  = "git"
  ssh_key   = "replace-with-private-key-material"
}
