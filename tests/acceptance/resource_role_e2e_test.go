package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRole_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_role.test"
	lookup := e2eReadByID("/roles")
	title := e2eRunPrefix + "-role"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_role", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleE2EConfig(title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", title),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

// testAccRoleE2EConfig takes its permission IDs from a role the tenant already
// has. RMS exposes permissions per role rather than as a global catalogue, and
// inventing IDs would make a rejected create look like a provider defect.
func testAccRoleE2EConfig(title string) string {
	return e2eCompanyConfig(e2eRunPrefix+"-role-company") + fmt.Sprintf(`
data "rms_roles" "all" {}

resource "rms_role" "test" {
  title          = %q
  description    = "E2E lifecycle check"
  company_id     = rms_company.test.id
  permission_ids = data.rms_roles.all.roles[0].permission_ids
}
`, title)
}
