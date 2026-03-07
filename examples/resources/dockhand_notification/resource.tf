resource "dockhand_notification" "example" {
  name = "apprise-example"
  type = "apprise"

  apprise_urls = [
    "json://",
  ]
}
