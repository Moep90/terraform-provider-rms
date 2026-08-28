data "teltonika-rms_company" "main" {
  id = 1
}

output "company_name" {
  value = data.teltonika-rms_company.main.company_name
}
