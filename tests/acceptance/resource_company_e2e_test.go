package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCompany_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_company.test"
	lookup := e2eReadByID("/companies")
	name := e2eRunPrefix + "-company"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_company", lookup),
		Steps: []resource.TestStep{
			{
				Config: e2eCompanyConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "company_name", name),
					e2eCheckExists(resourceName, lookup),
				),
			},
			{
				Config: e2eCompanyConfig(name + "-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "company_name", name+"-updated"),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}
