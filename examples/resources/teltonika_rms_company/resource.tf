resource "teltonika_rms_company" "main" {
  company_name = "Main Company"
}

resource "teltonika_rms_company" "subsidiary" {
  company_name = "Subsidiary Company"
  parent_id    = teltonika_rms_company.main.id
}
