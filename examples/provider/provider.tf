terraform {
  required_providers {
    rms = {
      source = "moep90/rms"
    }
  }
}

provider "rms" {
  # May also be set via the TELTONIKA_RMS_TOKEN environment variable.
  token = var.rms_token
}

variable "rms_token" {
  type      = string
  sensitive = true
}
