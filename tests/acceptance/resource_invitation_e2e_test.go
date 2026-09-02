package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInvitation_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_invitation.test"
	// RMS defines no GET on /users/invitations/{id}, so the invitation is
	// confirmed through its collection.
	lookup := e2eFindInList("/users/invitations")
	email := e2eRunPrefix + "-invite@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_invitation", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccInvitationE2EConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email", email),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

func testAccInvitationE2EConfig(email string) string {
	return e2eCompanyConfig(e2eRunPrefix+"-invitation-company") + fmt.Sprintf(`
resource "rms_invitation" "test" {
  email      = %q
  role       = "end_user"
  company_id = rms_company.test.id
}
`, email)
}
