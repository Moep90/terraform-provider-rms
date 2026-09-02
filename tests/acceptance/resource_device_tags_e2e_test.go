package acceptance

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceTags_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_device_tags.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_device_tags", e2eDeviceTagsLookup),
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceTagsE2EConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tag_ids.#", "1"),
					e2eCheckExists(resourceName, e2eDeviceTagsLookup),
				),
			},
		},
	})
}

// e2eDeviceTagsLookup confirms the device still carries every tag the
// assignment recorded. The assignment has no object of its own in RMS: it is
// read off the device, and a device that is gone carries nothing.
func e2eDeviceTagsLookup(ctx context.Context, client *api.Client, attrs map[string]string) error {
	var device map[string]interface{}
	if err := client.Get(ctx, "/devices/"+attrs["device_id"], nil, &device); err != nil {
		return err
	}

	assigned := map[string]bool{}
	if tags, ok := device["tags"].([]interface{}); ok {
		for _, tag := range tags {
			tagMap, ok := tag.(map[string]interface{})
			if !ok {
				continue
			}
			if id, ok := tagMap["id"].(float64); ok {
				assigned[strconv.FormatInt(int64(id), 10)] = true
			}
		}
	}

	for key, tagID := range attrs {
		if !strings.HasPrefix(key, "tag_ids.") || strings.HasSuffix(key, ".#") {
			continue
		}
		if !assigned[tagID] {
			return fmt.Errorf("device %s does not carry tag %s: %w", attrs["device_id"], tagID, api.ErrNotFound)
		}
	}

	return nil
}

func testAccDeviceTagsE2EConfig() string {
	return testAccDeviceE2EConfig(e2eRunPrefix+"-device-tags-device") + fmt.Sprintf(`
resource "rms_tag" "assigned" {
  name       = %q
  color      = "#00ff00"
  company_id = rms_company.test.id
}

resource "rms_device_tags" "test" {
  device_id = rms_device.test.id
  tag_ids   = [rms_tag.assigned.id]
}
`, e2eRunPrefix+"-device-tag")
}
