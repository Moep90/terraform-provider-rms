resource "teltonika-rms_company" "main" {
  company_name = "My Company"
}

resource "teltonika-rms_invitation" "new_user" {
  email      = "newuser@example.com"
  role       = "admin"
  company_id = teltonika-rms_company.main.id
}
