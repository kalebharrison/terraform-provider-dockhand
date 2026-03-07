resource "dockhand_settings_general" "example" {
  time_format             = "24h"
  date_format             = "YYYY-MM-DD"
  show_stopped_containers = true
  highlight_updates       = true
}
