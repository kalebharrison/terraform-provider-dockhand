resource "dockhand_network" "example" {
  name   = "tf-example-network"
  driver = "bridge"
  env    = "1"
}
