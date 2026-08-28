package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVPNHubRouteResource_CRUD(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	routeState := make(map[string]map[string]interface{})

	mux.HandleFunc("/vpn/hubs/routes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}

			ipAddress := ""
			if ip, ok := req["ip_address"].(string); ok {
				ipAddress = ip
			}
			netmask := ""
			if nm, ok := req["netmask"].(string); ok {
				netmask = nm
			}

			vpnHubID, _ := req["vpn_hub_id"].(float64)
			vpnHubUserID, _ := req["vpn_hub_user_id"].(float64)
			id := fmt.Sprintf("%d:%d:%s", int(vpnHubID), int(vpnHubUserID), ipAddress)

			routeState[id] = map[string]interface{}{
				"id":              id,
				"vpn_hub_id":      vpnHubID,
				"vpn_hub_user_id": vpnHubUserID,
				"ip_address":      ipAddress,
				"netmask":         netmask,
				"description":     req["description"],
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": id}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		if r.Method == http.MethodDelete {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}

			ipAddress := ""
			if ip, ok := req["ip_address"].(string); ok {
				ipAddress = ip
			}

			vpnHubID, _ := req["vpn_hub_id"].(float64)
			vpnHubUserID, _ := req["vpn_hub_user_id"].(float64)

			key := fmt.Sprintf("%d:%d:%s", int(vpnHubID), int(vpnHubUserID), ipAddress)
			delete(routeState, key)

			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/vpn/hubs/routes/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			id := parts[len(parts)-1]

			if r.Method == http.MethodGet {
				if cfg, ok := routeState[id]; ok {
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(cfg); err != nil {
						t.Logf("error encoding response: %v", err)
					}
					return
				}
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if r.Method == http.MethodPut {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Logf("error decoding request: %v", err)
				}
				if cfg, ok := routeState[id]; ok {
					if description, ok := req["description"].(string); ok {
						cfg["description"] = description
					}
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(cfg); err != nil {
						t.Logf("error encoding response: %v", err)
					}
					return
				}
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testVPNHubRouteConfig(server.URL, 1, 10, "192.168.1.0", "255.255.255.0", "Test route"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "vpn_hub_id", "1"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "vpn_hub_user_id", "10"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "ip_address", "192.168.1.0"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "netmask", "255.255.255.0"),
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "description", "Test route"),
				),
			},
			{
				Config: testVPNHubRouteConfig(server.URL, 1, 10, "192.168.1.0", "255.255.255.0", "Updated route"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_vpn_hub_route.test", "description", "Updated route"),
				),
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
