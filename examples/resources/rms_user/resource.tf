resource "rms_user" "admin" {
  username   = "admin"
  email      = "admin@example.com"
  role       = "admin"
  company_id = 12345
}
