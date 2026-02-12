resource "dockhand_volume" "example" {
  name   = "tf-example-volume"
  driver = "local"
  env    = "1"
}
