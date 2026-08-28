data "teltonika-rms_tags" "all" {}

output "tag_names" {
  value = data.teltonika-rms_tags.all.tags[*].name
}
