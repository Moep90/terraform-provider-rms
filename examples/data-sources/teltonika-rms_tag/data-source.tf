data "rms_tags" "all" {}

output "tag_names" {
  value = data.rms_tags.all.tags[*].name
}
