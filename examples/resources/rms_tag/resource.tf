resource "rms_tag" "production" {
  name       = "Production"
  color      = "#00ff00"
  company_id = 12345
}

resource "rms_tag" "development" {
  name       = "Development"
  color      = "#ff0000"
  company_id = 12345
}
