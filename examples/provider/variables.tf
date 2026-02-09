variable "dockhand_endpoint" {
  type        = string
  description = "Dockhand API endpoint."
}

variable "dockhand_username" {
  type        = string
  description = "Dockhand username."
}

variable "dockhand_password" {
  type        = string
  description = "Dockhand password."
  sensitive   = true
}

variable "dockhand_default_env" {
  type        = string
  description = "Default Dockhand environment ID."
  default     = "1"
}

variable "dockhand_user_password" {
  type        = string
  description = "Password used for the example user resource."
  sensitive   = true
}
