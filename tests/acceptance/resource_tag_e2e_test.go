package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTag_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_tag.test"
	lookup := e2eReadByID("/tags")
	name := e2eRunPrefix + "-tag"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_tag", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccTagE2EConfig(name, "#00ff00"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					e2eCheckExists(resourceName, lookup),
				),
			},
			{
				Config: testAccTagE2EConfig(name+"-updated", "#0000ff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "color", "#0000ff"),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

func testAccTagE2EConfig(name, color string) string {
	return e2eCompanyConfig(e2eRunPrefix+"-tag-company") + fmt.Sprintf(`
resource "rms_tag" "test" {
  name       = %q
  color      = %q
  company_id = rms_company.test.id
}
`, name, color)
}
