variable "dockhand_endpoint" {
  type        = string
  description = "Base URL for the Dockhand instance being bootstrapped."
}

variable "admin_username" {
  type        = string
  description = "Username for the first Dockhand administrator."
  default     = "admin"
}

variable "admin_password" {
  type        = string
  description = "Password for the first Dockhand administrator."
  sensitive   = true
}

variable "admin_email" {
  type        = string
  description = "Email for the first Dockhand administrator."
  default     = "admin@example.com"
}

variable "admin_display_name" {
  type        = string
  description = "Display name for the first Dockhand administrator."
  default     = "Dockhand Admin"
}
