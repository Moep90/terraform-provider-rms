package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCompanyResource(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1 for acceptance tests")
	}

	companyName := "Test Company"
	resourceName := "teltonika-rms_company.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCompanyConfig(companyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "company_name", companyName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func testAccCompanyConfig(name string) string {
	return fmt.Sprintf(`
provider "teltonika-rms" {
  token = os.Getenv("TELTONIKA_RMS_TOKEN")
}

resource "teltonika-rms_company" "test" {
  company_name = "%s"
}
`, name)
}
