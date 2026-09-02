package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDevice_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_device.test"
	lookup := e2eReadByID("/devices")
	name := e2eRunPrefix + "-device"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_device", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceE2EConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

func testAccDeviceE2EConfig(name string) string {
	return e2eCompanyConfig(name+"-company") + fmt.Sprintf(`
resource "rms_device" "test" {
  name               = %q
  device_series      = "rut"
  serial             = %q
  company_id         = rms_company.test.id
  auto_credit_enable = true
}
`, name, "E2E"+e2eRunID)
}
