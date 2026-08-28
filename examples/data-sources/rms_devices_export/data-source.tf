data "rms_devices_export" "all" {}

resource "local_file" "devices_csv" {
  content  = data.rms_devices_export.all.csv_data
  filename = "${path.module}/devices.csv"
}
