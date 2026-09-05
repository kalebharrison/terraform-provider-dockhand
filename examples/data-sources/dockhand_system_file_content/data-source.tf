data "dockhand_system_file_content" "compose" {
  # Prefer stack/data paths. Dockhand blocks /etc, /proc, /root and secret trees.
  path = "/docker/stacks/myapp/compose.yaml"
}
