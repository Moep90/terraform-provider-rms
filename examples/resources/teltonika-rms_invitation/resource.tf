resource "rms_company" "main" {
  company_name = "My Company"
}

resource "rms_invitation" "new_user" {
  email      = "newuser@example.com"
  role       = "admin"
  company_id = rms_company.main.id
}
