package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEmailConfiguration_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_email_configuration.test"
	// RMS defines no GET on /email-configurations/{id}, so the object is
	// confirmed through its collection.
	lookup := e2eFindInList("/email-configurations")
	name := e2eRunPrefix + "-smtp"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_email_configuration", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailConfigurationE2EConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

func testAccEmailConfigurationE2EConfig(name string) string {
	return fmt.Sprintf(`
resource "rms_email_configuration" "test" {
  name     = %q
  host     = "smtp.example.com"
  port     = 587
  email    = %q
  username = %q
  password = "e2e-placeholder"
}
`, name, e2eRunPrefix+"@example.com", e2eRunPrefix)
}
