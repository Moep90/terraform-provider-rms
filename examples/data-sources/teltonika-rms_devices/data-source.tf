data "teltonika-rms_devices" "all" {}

data "teltonika-rms_devices" "online" {
  company_id = 12345
  status     = "online"
}

output "device_names" {
  value = data.teltonika-rms_devices.all.devices[*].name
}
