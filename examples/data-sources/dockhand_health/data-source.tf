data "dockhand_health" "current" {
  env = "1"
}

output "dockhand_status" {
  value = data.dockhand_health.current.status
}
