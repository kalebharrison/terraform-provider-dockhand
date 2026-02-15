resource "dockhand_stack_env" "example" {
  env        = "2"
  stack_name = "my-stack"

  raw_content = <<-EOT
APP_ENV=prod
EOT

  secret_variables = [
    {
      key       = "API_TOKEN"
      value     = "example-secret"
      is_secret = true
    }
  ]
}
