resource "dockhand_prune_action" "example" {
  env                 = "1"
  mode                = "containers"
  wait_for_completion = false
  trigger             = "run-1"
}
