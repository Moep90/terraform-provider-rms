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
)

func TestVPNHubResource_CRUD(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	vpnHubState := make(map[int]map[string]interface{})
	nextID := 1

	mux.HandleFunc("/vpn/hubs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}

			currentID := nextID
			nextID++

			name := ""
			if n, ok := req["name"].(string); ok {
				name = n
			}
			description := ""
			if d, ok := req["description"].(string); ok {
				description = d
			}
			var tagIDs []interface{}
			if t, ok := req["tag_id"].([]interface{}); ok {
				tagIDs = t
			}

			vpnHubState[currentID] = map[string]interface{}{
				"id":          float64(currentID),
				"name":        name,
				"description": description,
				"company_id":  float64(1),
				"hub_zone":    req["hub_zone"],
				"vpn_type":    req["vpn_type"],
				"tag_id":      tagIDs,
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": float64(currentID)}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		if r.Method == http.MethodDelete {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}
			if idList, ok := req["id"].([]interface{}); ok {
				for _, id := range idList {
					if f, ok := id.(float64); ok {
						delete(vpnHubState, int(f))
					}
				}
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/vpn/hubs/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			id, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if r.Method == http.MethodGet {
				if cfg, ok := vpnHubState[id]; ok {
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
				if cfg, ok := vpnHubState[id]; ok {
					if description, ok := req["description"].(string); ok {
						cfg["description"] = description
					}
					if tagIDs, ok := req["tag_id"].([]interface{}); ok {
						cfg["tag_id"] = tagIDs
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
				Config: testVPNHubConfig(server.URL, "Test VPN Hub", "Test description", 1, "frankfurt-1", "tun", []int{10, 20}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_vpn_hub.test", "name", "Test VPN Hub"),
					resource.TestCheckResourceAttr("rms_vpn_hub.test", "description", "Test description"),
					resource.TestCheckResourceAttr("rms_vpn_hub.test", "company_id", "1"),
					resource.TestCheckResourceAttr("rms_vpn_hub.test", "hub_zone", "frankfurt-1"),
					resource.TestCheckResourceAttr("rms_vpn_hub.test", "vpn_type", "tun"),
					resource.TestCheckTypeSetElemAttr("rms_vpn_hub.test", "tag_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_vpn_hub.test", "tag_ids.*", "20"),
				),
			},
			{
				Config: testVPNHubConfig(server.URL, "Test VPN Hub", "Updated description", 1, "frankfurt-1", "tun", []int{10, 20}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_vpn_hub.test", "name", "Test VPN Hub"),
					resource.TestCheckResourceAttr("rms_vpn_hub.test", "description", "Updated description"),
					resource.TestCheckTypeSetElemAttr("rms_vpn_hub.test", "tag_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_vpn_hub.test", "tag_ids.*", "20"),
				),
			},
		},
	})
}

func testVPNHubConfig(baseURL string, name, description string, companyID int64, hubZone, vpnType string, tagIDs []int) string {
	tagStr := "["
	for i, id := range tagIDs {
		if i > 0 {
			tagStr += ", "
		}
		tagStr += fmt.Sprintf("%d", id)
	}
	tagStr += "]"

	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_vpn_hub" "test" {
  name         = "%s"
  description  = "%s"
  company_id   = %d
  hub_zone     = "%s"
  vpn_type     = "%s"
  tag_ids      = %s
}
`, baseURL, name, description, companyID, hubZone, vpnType, tagStr)
}
