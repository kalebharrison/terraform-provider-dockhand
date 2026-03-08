data "dockhand_registry_tags" "example" {
  image    = "library/busybox"
  registry = "1"
  page     = 1
  page_size = 25
}
