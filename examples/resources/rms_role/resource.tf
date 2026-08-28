resource "rms_role" "admin" {
  title          = "Administrator"
  description    = "Full access role"
  company_id     = 123456
  permission_ids = [1, 2, 3, 4, 5]
}

resource "rms_role" "viewer" {
  title          = "Viewer"
  description    = "Read-only access"
  company_id     = 123456
  permission_ids = [1]
}
