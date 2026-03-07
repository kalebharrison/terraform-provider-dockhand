resource "dockhand_git_repository" "example" {
  name           = "stacks"
  url            = "https://github.com/example/stacks.git"
  branch         = "main"
  credential_id  = "1"
  environment_id = "2"

  auto_update          = false
  auto_update_schedule = "daily"
  auto_update_cron     = "0 3 * * *"
  webhook_enabled      = false
}
