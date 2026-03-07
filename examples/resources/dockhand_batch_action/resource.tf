resource "dockhand_batch_action" "example" {
  env                 = "2"
  entity_type         = "containers"
  operation           = "restart"
  item_ids            = ["replace-with-container-id"]
  wait_for_completion = true
  timeout_seconds     = 30
  poll_interval_ms    = 500
  trigger             = "example-run-1"
}
