variable "dockhand_endpoint" {
  type        = string
  description = "Base URL for the Dockhand API."
}

variable "dockhand_username" {
  type        = string
  description = "Dockhand login username."
}

variable "dockhand_password" {
  type        = string
  description = "Dockhand login password."
  sensitive   = true
}

variable "dockhand_default_env" {
  type        = string
  description = "Dockhand environment ID that will receive the Git-backed stack."
}

variable "git_username" {
  type        = string
  description = "Git provider username or account name."
}

variable "git_token" {
  type        = string
  description = "Git access token for the Dockhand Git credential."
  sensitive   = true
}

variable "git_repository_url" {
  type        = string
  description = "HTTPS Git repository URL."
}

variable "git_branch" {
  type        = string
  description = "Repository branch used by Dockhand."
  default     = "main"
}

variable "stack_name" {
  type        = string
  description = "Dockhand Git stack name."
  default     = "example-git-stack"
}

variable "compose_path" {
  type        = string
  description = "Path to the compose file within the Git repository."
  default     = "stacks/example/compose.yaml"
}
