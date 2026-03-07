resource "dockhand_registry" "example" {
  name       = "Docker Hub"
  url        = "https://registry.hub.docker.com"
  is_default = true
}
