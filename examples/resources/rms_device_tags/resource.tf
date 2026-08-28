resource "rms_device_tags" "main" {
  device_id = rms_device.main.id
  tag_ids   = [rms_tag.production.id, rms_tag.critical.id]
}
