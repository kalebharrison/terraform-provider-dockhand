resource "dockhand_container_file" "example" {
  env          = "2"
  container_id = "abc123"
  path         = "/tmp/example.txt"
  content      = "hello from terraform"
}
