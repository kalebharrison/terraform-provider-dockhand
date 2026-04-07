terraform {
  required_providers {
    dockhand = {
      source  = "kalebharrison/dockhand"
      version = ">= 0.1.63"
    }
  }
}

provider "dockhand" {
  endpoint    = var.dockhand_endpoint
  username    = var.dockhand_username
  password    = var.dockhand_password
  default_env = var.dockhand_default_env
}

resource "dockhand_git_credential" "github_pat" {
  name     = "github-pat"
  provider = "github"
  username = var.git_username
  token    = var.git_token
}

resource "dockhand_git_repository" "stacks_repo" {
  name          = "platform-stacks"
  url           = var.git_repository_url
  branch        = var.git_branch
  credential_id = dockhand_git_credential.github_pat.id
  environment_id = var.dockhand_default_env
}

resource "dockhand_git_stack" "app" {
  name          = var.stack_name
  repository_id = dockhand_git_repository.stacks_repo.id
  environment_id = var.dockhand_default_env
  compose_path  = var.compose_path
  auto_deploy   = true
  webhook_enabled = false
}

resource "dockhand_git_stack_env_file" "app_env" {
  git_stack_id = dockhand_git_stack.app.id
  file_name    = ".env"
  content      = <<-ENV
    APP_ENV=production
    TZ=UTC
  ENV
}
