resource "rms_company" "main" {
  company_name = "My Company"
}

resource "rms_device" "router" {
  name               = "Office Router"
  device_series      = "rut"
  serial             = "0123456789"
  mac                = "00:11:22:33:44:55"
  company_id         = rms_company.main.id
  auto_credit_enable = true
  password           = "device-password"
}
