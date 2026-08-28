package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTask_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	token := os.Getenv("RMS_ADMIN_TOKEN")
	baseURL := os.Getenv("RMS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://rms.teltonika-networks.com/api"
	}

	resourceName := "teltonika-rms_task.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTaskE2EConfig(baseURL, token, "E2E Test Task", "reboot", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "E2E Test Task"),
					resource.TestCheckResourceAttr(resourceName, "task_type", "reboot"),
					resource.TestCheckResourceAttr(resourceName, "company_id", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
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

func testAccTaskE2EConfig(baseURL, token, name, taskType string, companyID int) string {
	return fmt.Sprintf(`
provider "teltonika-rms" {
  token     = %q
  base_url  = %q
}

resource "teltonika-rms_task" "test" {
  name        = %q
  task_type   = %q
  company_id  = %d
  payload     = "{\"command\":\"reboot\"}"
}
`, token, baseURL, name, taskType, companyID)
}
