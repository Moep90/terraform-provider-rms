data "rms_device" "main" {
  id = 1
}

output "device_serial" {
  value = data.rms_device.main.serial
}
