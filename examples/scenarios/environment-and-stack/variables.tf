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
  description = "Default environment ID used by the provider when env is omitted."
}

variable "environment_name" {
  type        = string
  description = "Name for the managed Dockhand environment."
  default     = "docker-socket-local"
}
