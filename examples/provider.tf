provider "teltonika-rms" {
  token = var.rms_token
  base_url = var.rms_base_url
}

variable "rms_token" {
  type = string
  sensitive = true
}

variable "rms_base_url" {
  default = "https://rms.teltonika-networks.com/api"
}
