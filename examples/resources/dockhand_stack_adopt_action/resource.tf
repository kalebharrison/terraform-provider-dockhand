resource "dockhand_stack_adopt_action" "example" {
  environment_id = 2
  trigger        = "manual-1"

  stacks = [
    {
      name         = "example"
      compose_path = "/app/data/stacks/example/compose.yaml"
    }
  ]
}
