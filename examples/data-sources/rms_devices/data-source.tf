data "rms_devices" "all" {}

data "rms_devices" "online" {
  company_id = 12345
  status     = "online"
}

output "device_names" {
  value = data.rms_devices.all.devices[*].name
}
