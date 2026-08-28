data "teltonika-rms_companies" "all" {}

output "company_names" {
  value = data.teltonika-rms_companies.all.companies[*].company_name
}
