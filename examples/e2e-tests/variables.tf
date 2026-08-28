variable "rms_token" {
  description = "RMS API token for authentication"
  type        = string
  sensitive   = true
}

variable "rms_base_url" {
  description = "RMS API base URL (e.g., https://eu.rms.teltonika.lt/api)"
  type        = string
  default     = "https://eu.rms.teltonika.lt/api"
}
