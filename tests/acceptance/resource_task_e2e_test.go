package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTask_E2E pins the apply-time rejection against real RMS. There is no
// create operation for an individual task, so there is no lifecycle to run and
// no object to check for: the whole contract is that apply fails and nothing is
// left behind.
func TestAccTask_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccTaskE2EConfig(e2eRunPrefix + "-task"),
				ExpectError: regexp.MustCompile(`rms_task cannot be created`),
			},
		},
	})
}

func testAccTaskE2EConfig(name string) string {
	return e2eCompanyConfig(e2eRunPrefix+"-task-company") + fmt.Sprintf(`
resource "rms_task" "test" {
  name       = %q
  task_type  = "reboot"
  company_id = rms_company.test.id
  payload    = "{\"command\":\"reboot\"}"
}
`, name)
}
