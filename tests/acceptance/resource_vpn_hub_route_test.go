package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// vpnHubRouteMux serves the route operations RMS defines: GET and POST on
// /vpn/hubs/routes and DELETE on /vpn/hubs/routes/{vpn_hub_id}. There is no
// read-by-id and no update, so those paths 404.
func vpnHubRouteMux(t *testing.T, state map[string]map[string]interface{}) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	nextID := 1

	routeKey := func(hubID, userID int, ip string) string {
		return fmt.Sprintf("%d:%d:%s", hubID, userID, ip)
	}

	mux.HandleFunc("/vpn/hubs/routes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			hubID := r.URL.Query().Get("vpn_hub_id")
			userID := r.URL.Query().Get("vpn_hub_user_id")
			if hubID == "" || userID == "" {
				t.Errorf("GET /vpn/hubs/routes must be scoped by hub and user, got %q", r.URL.RawQuery)
			}

			routes := make([]interface{}, 0, len(state))
			for key, route := range state {
				if strings.HasPrefix(key, hubID+":"+userID+":") {
					routes = append(routes, route)
				}
			}
			writeRMSList(t, w, routes)

		case http.MethodPost:
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}

			ip, _ := req["ip_address"].(string)
			netmask, _ := req["netmask"].(string)
			hubID, _ := req["vpn_hub_id"].(float64)
			userID, _ := req["vpn_hub_user_id"].(float64)
			description, _ := req["description"].(string)

			key := routeKey(int(hubID), int(userID), ip)
			state[key] = map[string]interface{}{
				"id":      float64(nextID),
				"hub_id":  hubID,
				"ip":      ip,
				"netmask": netmask,
				"name":    description,
			}
			nextID++

			w.Header().Set("Content-Type", "application/json")
			// The provider composes the resource id itself from hub, user and
			// address, so the create response only has to succeed.
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": key}); err != nil {
				t.Logf("error encoding response: %v", err)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	mux.HandleFunc("/vpn/hubs/routes/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		hubID, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// The route is selected through the request body, not the path.
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("DELETE /vpn/hubs/routes/{id} must carry a JSON body: %v", err)
		}
		ip, _ := req["ip_address"].(string)
		userID, _ := req["vpn_hub_user_id"].(float64)
		if ip == "" {
			t.Errorf("DELETE /vpn/hubs/routes/{id} body has no ip_address: %v", req)
		}

		delete(state, routeKey(hubID, int(userID), ip))
		w.WriteHeader(http.StatusOK)
	})

	return mux
}

func TestVPNHubRouteResource_CRUD(t *testing.T) {
	routeState := make(map[string]map[string]interface{})
	server := httptest.NewServer(vpnHubRouteMux(t, routeState))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testVPNHubRouteConfig(server.URL, 1, 10, "192.168.1.0", "255.255.255.0", "Test route"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_vpn_hub_route.test", plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "vpn_hub_id", "1"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "vpn_hub_user_id", "10"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "ip_address", "192.168.1.0"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "netmask", "255.255.255.0"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "description", "Test route"),
				),
			},
			{
				// RMS exposes no update for routes, so changing description has
				// to be planned as a replacement.
				Config: testVPNHubRouteConfig(server.URL, 1, 10, "192.168.1.0", "255.255.255.0", "Updated route"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_vpn_hub_route.test", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "description", "Updated route"),
			},
		},
	})
}

// TestVPNHubRouteResource_ReadRemovesDeletedRoute covers the read-from-list
// rule: a route deleted out of band drops out of /vpn/hubs/routes, so it must
// leave state and the next plan must show a create.
func TestVPNHubRouteResource_ReadRemovesDeletedRoute(t *testing.T) {
	routeState := make(map[string]map[string]interface{})
	server := httptest.NewServer(vpnHubRouteMux(t, routeState))
	defer server.Close()

	config := testVPNHubRouteConfig(server.URL, 1, 10, "192.168.1.0", "255.255.255.0", "Test route")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "ip_address", "192.168.1.0"),
			},
			{
				PreConfig: func() {
					for key := range routeState {
						delete(routeState, key)
					}
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_vpn_hub_route.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testVPNHubRouteConfig(baseURL string, vpnHubID, vpnHubUserID int64, ipAddress, netmask, description string) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_vpn_hub_route" "test" {
  vpn_hub_id       = %d
  vpn_hub_user_id  = %d
  ip_address       = "%s"
  netmask          = "%s"
  description      = "%s"
}
`, baseURL, vpnHubID, vpnHubUserID, ipAddress, netmask, description)
}
