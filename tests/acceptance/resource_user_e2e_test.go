package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUser_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	token := os.Getenv("RMS_ADMIN_TOKEN")
	baseURL := os.Getenv("RMS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://rms.teltonika-networks.com/api"
	}

	resourceName := "rms_user.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccUserE2EConfig(baseURL, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "username"),
					resource.TestCheckResourceAttrSet(resourceName, "email"),
					resource.TestCheckResourceAttrSet(resourceName, "role"),
				),
			},
			{
				Config: testAccUserE2EConfigUpdated(baseURL, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "e2e-user-updated"),
					resource.TestCheckResourceAttr(resourceName, "email", "updated@example.com"),
					resource.TestCheckResourceAttr(resourceName, "role", "admin"),
				),
			},
		},
	})
}

func testAccUserE2EConfig(baseURL, token string) string {
	return fmt.Sprintf(`
provider "rms" {
  token     = %q
  base_url  = %q
}

resource "rms_company" "test" {
  company_name = "E2E User Test Company"
}

resource "rms_user" "test" {
  username   = "e2e-user-test"
  email      = "e2e@example.com"
  role       = "viewer"
  company_id = rms_company.test.id
}
`, token, baseURL)
}

func testAccUserE2EConfigUpdated(baseURL, token string) string {
	return fmt.Sprintf(`
provider "rms" {
  token     = %q
  base_url  = %q
}

resource "rms_company" "test" {
  company_name = "E2E User Test Company"
}

resource "rms_user" "test" {
  username   = "e2e-user-updated"
  email      = "updated@example.com"
  role       = "admin"
  company_id = rms_company.test.id
}
`, token, baseURL)
}
