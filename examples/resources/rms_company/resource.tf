resource "rms_company" "main" {
  company_name = "Main Company"
}

resource "rms_company" "subsidiary" {
  company_name = "Subsidiary Company"
  parent_id    = rms_company.main.id
}
