package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccVPNHub_E2E is expected to fail against real RMS: POST /vpn/hubs is an
// asynchronous Status API operation and answers with a channel handle, not an
// id, while Create parses result["id"]. See deferred-work.md. The failure is
// the finding, so this test must not be skipped or weakened to go green.
func TestAccVPNHub_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_vpn_hub.test"
	// RMS defines no GET on /vpn/hubs/{id}, so the hub is confirmed through its
	// collection.
	lookup := e2eFindInList("/vpn/hubs", "company_id")
	name := e2eRunPrefix + "-vpn-hub"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_vpn_hub", lookup),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNHubE2EConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					e2eCheckExists(resourceName, lookup),
				),
			},
		},
	})
}

func testAccVPNHubE2EConfig(name string) string {
	return e2eCompanyConfig(e2eRunPrefix+"-vpn-hub-company") + fmt.Sprintf(`
resource "rms_vpn_hub" "test" {
  name        = %q
  description = "E2E lifecycle check"
  company_id  = rms_company.test.id
  hub_zone    = %q
  vpn_type    = "tun"
}
`, name, e2eVPNHubZone())
}

// e2eVPNHubZone is the zone an E2E VPN hub is created in. Zones are tenant and
// region specific, so the value comes from the environment.
func e2eVPNHubZone() string {
	if v := os.Getenv("RMS_VPN_HUB_ZONE"); v != "" {
		return v
	}
	return "frankfurt-1"
}
