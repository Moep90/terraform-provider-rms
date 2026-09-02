package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUser_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_user.test"
	lookup := e2eReadByID("/users")
	username := e2eRunPrefix + "-user"
	email := e2eRunPrefix + "@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_user", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccUserE2EConfig(username, email, "end_user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email", email),
					e2eCheckExists(resourceName, lookup),
				),
			},
			{
				Config: testAccUserE2EConfig(username+"-updated", email, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", username+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "role", "admin"),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

func testAccUserE2EConfig(username, email, role string) string {
	return e2eCompanyConfig(e2eRunPrefix+"-user-company") + fmt.Sprintf(`
resource "rms_user" "test" {
  username   = %q
  email      = %q
  role       = %q
  company_id = rms_company.test.id
}
`, username, email, role)
}
