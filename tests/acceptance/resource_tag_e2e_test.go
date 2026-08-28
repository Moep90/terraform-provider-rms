package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTag_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	token := os.Getenv("RMS_ADMIN_TOKEN")
	baseURL := os.Getenv("RMS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://rms.teltonika-networks.com/api"
	}

	resourceName := "rms_tag.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTagE2EConfig(baseURL, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "color"),
				),
			},
			{
				Config: testAccTagE2EConfigUpdated(baseURL, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "e2e-tag-updated"),
					resource.TestCheckResourceAttr(resourceName, "color", "#0000ff"),
				),
			},
		},
	})
}

func testAccTagE2EConfig(baseURL, token string) string {
	return fmt.Sprintf(`
provider "rms" {
  token     = %q
  base_url  = %q
}

resource "rms_company" "test" {
  company_name = "E2E Tag Test Company"
}

resource "rms_tag" "test" {
  name       = "e2e-tag-test"
  color      = "#00ff00"
  company_id = rms_company.test.id
}
`, token, baseURL)
}

func testAccTagE2EConfigUpdated(baseURL, token string) string {
	return fmt.Sprintf(`
provider "rms" {
  token     = %q
  base_url  = %q
}

resource "rms_company" "test" {
  company_name = "E2E Tag Test Company"
}

resource "rms_tag" "test" {
  name       = "e2e-tag-updated"
  color      = "#0000ff"
  company_id = rms_company.test.id
}
`, token, baseURL)
}
