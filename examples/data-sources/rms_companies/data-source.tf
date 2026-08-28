data "rms_companies" "all" {}

output "company_names" {
  value = data.rms_companies.all.companies[*].company_name
}
