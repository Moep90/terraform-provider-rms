data "teltonika_rms_devices" "all" {}

data "teltonika_rms_devices" "online" {
  company_id = 12345
  status     = "online"
}

output "device_names" {
  value = data.teltonika_rms_devices.all.devices[*].name
}
