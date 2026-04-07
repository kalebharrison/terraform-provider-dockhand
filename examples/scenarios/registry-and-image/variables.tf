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
  description = "Dockhand environment ID used for image operations."
}

variable "registry_username" {
  type        = string
  description = "Registry username for Dockhand registry auth."
}

variable "registry_password" {
  type        = string
  description = "Registry password or token."
  sensitive   = true
}

variable "image_name" {
  type        = string
  description = "Fully-qualified image reference to manage."
  default     = "ghcr.io/example/app:latest"
}
