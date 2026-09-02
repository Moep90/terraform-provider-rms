package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTaskGroup_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_task_group.test"
	lookup := e2eReadByID("/devices/tasks/groups")
	name := e2eRunPrefix + "-task-group"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_task_group", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskGroupE2EConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

func testAccTaskGroupE2EConfig(name string) string {
	return e2eCompanyConfig(e2eRunPrefix+"-task-group-company") + fmt.Sprintf(`
resource "rms_task_group" "test" {
  name       = %q
  company_id = rms_company.test.id
}
`, name)
}
