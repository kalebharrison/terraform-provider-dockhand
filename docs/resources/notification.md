# dockhand_notification

Manages a Dockhand notification integration via `/api/notifications`.

## Example Usage

### Apprise

```terraform
resource "dockhand_notification" "apprise" {
  name = "apprise"
  type = "apprise"

  apprise_urls = [
    "json://",
  ]
}
```

### SMTP

```terraform
resource "dockhand_notification" "smtp" {
  name = "smtp"
  type = "smtp"

  smtp_host      = "smtp.example.invalid"
  smtp_port      = 587
  smtp_from_email = "dockhand@example.invalid"
  smtp_to_emails = [
    "ops@example.invalid",
  ]

  smtp_username = "user"
  smtp_password = "pass"

  smtp_use_tls         = true
  smtp_starttls        = true
  smtp_skip_tls_verify = true
}
```

## Notes

- Dockhand returns `eventTypes` and may default them to a large set on create. If you omit `event_types` in Terraform, the resource will adopt Dockhand's defaults on create and then store the resulting set in state.
- `smtp_password` is marked sensitive. Terraform will store it in state if set (ensure your state is secured).

