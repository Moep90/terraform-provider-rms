package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTaskGroup_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	token := os.Getenv("RMS_ADMIN_TOKEN")
	baseURL := os.Getenv("RMS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://rms.teltonika-networks.com/api"
	}

	resourceName := "rms_task_group.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTaskGroupE2EConfig(baseURL, token, "E2E Test Task Group", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "E2E Test Task Group"),
					resource.TestCheckResourceAttr(resourceName, "company_id", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "task_count"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceName].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTaskGroupE2EConfig(baseURL, token, name string, companyID int) string {
	return fmt.Sprintf(`
provider "rms" {
  token     = %q
  base_url  = %q
}

resource "rms_task_group" "test" {
  name        = %q
  company_id  = %d
}
`, token, baseURL, name, companyID)
}
