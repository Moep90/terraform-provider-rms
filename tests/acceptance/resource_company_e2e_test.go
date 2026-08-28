package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCompany_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	token := os.Getenv("RMS_ADMIN_TOKEN")
	baseURL := os.Getenv("RMS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://rms.teltonika-networks.com/api"
	}

	resourceName := "teltonika-rms_company.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCompanyE2EConfig(baseURL, token, "E2E Test Company"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "company_name"),
				),
			},
			{
				Config: testAccCompanyE2EConfig(baseURL, token, "E2E Test Company Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "company_name", "E2E Test Company Updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) { return s.RootModule().Resources[resourceName].Primary.ID, nil },
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCompanyE2EConfig(baseURL, token, name string) string {
	return fmt.Sprintf(`
provider "teltonika-rms" {
  token     = %q
  base_url  = %q
}

resource "teltonika-rms_company" "test" {
  company_name = %q
}
`, token, baseURL, name)
}
