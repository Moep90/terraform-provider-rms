package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAlertConfiguration_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_alert_configuration.test"
	lookup := e2eReadByID("/alerts-configurations")
	subject := e2eRunPrefix + "-alert"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_alert_configuration", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccAlertConfigurationE2EConfig(subject),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "subject", subject),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

// testAccAlertConfigurationE2EConfig creates the device the alert configuration
// hangs off, rather than attaching to a device the tenant already has.
func testAccAlertConfigurationE2EConfig(subject string) string {
	return testAccDeviceE2EConfig(e2eRunPrefix+"-alert-device") + fmt.Sprintf(`
resource "rms_alert_configuration" "test" {
  device_id     = rms_device.test.id
  alert_type_id = 5
  action        = 1
  subject       = %q
  message       = "E2E lifecycle check"
  email         = %q
}
`, subject, e2eRunPrefix+"@example.com")
}
