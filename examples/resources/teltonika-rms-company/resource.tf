resource "teltonika-rms_company" "main" {
  company_name = "Main Company"
}

resource "teltonika-rms_company" "subsidiary" {
  company_name = "Subsidiary Company"
  parent_id    = teltonika-rms_company.main.id
}
