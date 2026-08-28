data "rms_device_esim_bootstrap" "main" {
  device_id = rms_device.main.id
}

output "esim_bootstrap_status" {
  value = data.rms_device_esim_bootstrap.main.status
}
