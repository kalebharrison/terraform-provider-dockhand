resource "dockhand_notification_test_action" "smtp" {
  type = "smtp"
  config_json = jsonencode({
    host       = "smtp.example.local"
    port       = 25
    from_email = "dockhand@example.local"
    to_emails  = ["ops@example.local"]
  })

  fail_on_error = false
  trigger       = "run-1"
}
