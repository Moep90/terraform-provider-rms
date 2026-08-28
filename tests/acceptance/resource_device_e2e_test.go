package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDevice_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	token := os.Getenv("RMS_ADMIN_TOKEN")
	baseURL := os.Getenv("RMS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://rms.teltonika-networks.com/api"
	}

	resourceName := "teltonika-rms_device.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceE2EConfig(baseURL, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "device_series"),
					resource.TestCheckResourceAttrSet(resourceName, "serial"),
				),
			},
		},
	})
}

func testAccDeviceE2EConfig(baseURL, token string) string {
	return fmt.Sprintf(`
provider "teltonika-rms" {
  token     = %q
  base_url  = %q
}

resource "teltonika-rms_company" "test" {
  company_name = "E2E Device Test Company"
}

resource "teltonika-rms_device" "test" {
  name             = "e2e-device-test"
  device_series    = "rut"
  serial           = "E2E123456789"
  company_id       = teltonika-rms_company.test.id
  auto_credit_enable = true
}
`, token, baseURL)
}
