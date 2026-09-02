package acceptance

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccVPNHubRoute_E2E is expected to fail against real RMS for the same
// reason as TestAccVPNHub_E2E: POST /vpn/hubs/routes answers with a Status API
// channel handle rather than an id, while Create parses result["id"]. See
// deferred-work.md. The failure is the finding.
func TestAccVPNHubRoute_E2E(t *testing.T) {
	if !isRealAPITest() {
		t.Skip("RMS_ADMIN_TOKEN not set, skipping real API test")
	}

	const resourceName = "rms_vpn_hub_route.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             e2eCheckDestroyed("rms_vpn_hub_route", e2eVPNHubRouteLookup),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNHubRouteE2EConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.0.0"),
					e2eCheckExists(resourceName, e2eVPNHubRouteLookup),
				),
			},
		},
	})
}

// e2eVPNHubRouteLookup confirms the route is still listed for its hub. RMS
// defines no GET on /vpn/hubs/routes/{id}, and the collection identifies a
// route by address rather than by id.
func e2eVPNHubRouteLookup(ctx context.Context, client *api.Client, attrs map[string]string) error {
	var routes []map[string]interface{}
	err := client.Get(ctx, "/vpn/hubs/routes", map[string]string{
		"vpn_hub_id":      attrs["vpn_hub_id"],
		"vpn_hub_user_id": attrs["vpn_hub_user_id"],
	}, &routes)
	if err != nil {
		return err
	}

	for _, route := range routes {
		ip, ok := route["ip"].(string)
		if !ok {
			ip, ok = route["ip_address"].(string)
		}
		if !ok {
			continue
		}

		netmask, ok := route["netmask"].(string)
		if !ok {
			continue
		}

		if ip == attrs["ip_address"] && netmask == attrs["netmask"] {
			return nil
		}
	}

	return api.ErrNotFound
}

func testAccVPNHubRouteE2EConfig() string {
	return testAccVPNHubE2EConfig(e2eRunPrefix+"-route-hub") + fmt.Sprintf(`
resource "rms_vpn_hub_route" "test" {
  vpn_hub_id      = rms_vpn_hub.test.id
  vpn_hub_user_id = %s
  ip_address      = "10.0.0.0"
  netmask         = "255.255.255.0"
  description     = "E2E lifecycle check"
}
`, e2eVPNHubUserID())
}

// e2eVPNHubUserID is the hub user a route is created for. RMS exposes no
// Terraform resource for hub users, so the id is tenant specific and comes from
// the environment.
func e2eVPNHubUserID() string {
	if v := os.Getenv("RMS_VPN_HUB_USER_ID"); v != "" {
		return v
	}
	return "0"
}
