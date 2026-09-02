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

// vpnHubMux serves the VPN hub operations RMS actually defines: GET, POST and
// DELETE on /vpn/hubs and PUT on /vpn/hubs/{id}. Every other route 404s, so a
// call to an undefined path fails the test.
func vpnHubMux(t *testing.T, state map[int]map[string]interface{}) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	nextID := 1

	mux.HandleFunc("/vpn/hubs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			hubs := make([]interface{}, 0, len(state))
			for _, hub := range state {
				hubs = append(hubs, hub)
			}
			writeRMSList(t, w, hubs)

		case http.MethodPost:
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}

			currentID := nextID
			nextID++

			name, _ := req["name"].(string)
			description, _ := req["description"].(string)

			state[currentID] = map[string]interface{}{
				"id":          float64(currentID),
				"name":        name,
				"description": description,
				"company_id":  float64(1),
				"hub_zone":    req["hub_zone"],
				"vpn_type":    req["vpn_type"],
				"tags":        tagObjects(req["tag_id"]),
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": float64(currentID)}); err != nil {
				t.Logf("error encoding response: %v", err)
			}

		case http.MethodDelete:
			// RMS selects the hubs to delete through the request body.
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("DELETE /vpn/hubs must carry a JSON body: %v", err)
			}
			idList, ok := req["id"].([]interface{})
			if !ok {
				t.Errorf("DELETE /vpn/hubs body has no id list: %v", req)
			}
			for _, id := range idList {
				if f, ok := id.(float64); ok {
					delete(state, int(f))
				}
			}
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	mux.HandleFunc("/vpn/hubs/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// RMS defines PUT here and nothing else; notably there is no GET.
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Logf("error decoding request: %v", err)
		}
		hub, ok := state[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if description, ok := req["description"].(string); ok {
			hub["description"] = description
		}
		if _, ok := req["tag_id"]; ok {
			hub["tags"] = tagObjects(req["tag_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(hub); err != nil {
			t.Logf("error encoding response: %v", err)
		}
	})

	return mux
}

// writeRMSList emits the envelope RMS wraps its collections in.
func writeRMSList(t *testing.T, w http.ResponseWriter, items []interface{}) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	body := map[string]interface{}{
		"success": true,
		"data":    items,
		"meta":    map[string]interface{}{"total": len(items)},
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Logf("error encoding response: %v", err)
	}
}

// tagObjects turns the tag_id array a request carries into the tag objects RMS
// reports back on the hub record.
func tagObjects(raw interface{}) []interface{} {
	ids, ok := raw.([]interface{})
	if !ok {
		return []interface{}{}
	}

	tags := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		tags = append(tags, map[string]interface{}{"id": id})
	}
	return tags
}

func TestVPNHubResource_CRUD(t *testing.T) {
	vpnHubState := make(map[int]map[string]interface{})
	server := httptest.NewServer(vpnHubMux(t, vpnHubState))
	defer server.Close()

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

// TestVPNHubResource_ReadRemovesDeletedHub covers the read-from-list rule: a
// hub deleted out of band drops out of /vpn/hubs, so it must leave state and
// the next plan must show a create.
func TestVPNHubResource_ReadRemovesDeletedHub(t *testing.T) {
	vpnHubState := make(map[int]map[string]interface{})
	server := httptest.NewServer(vpnHubMux(t, vpnHubState))
	defer server.Close()

	config := testVPNHubConfig(server.URL, "Test VPN Hub", "Test description", 1, "frankfurt-1", "tun", []int{10})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("rms_vpn_hub.test", "name", "Test VPN Hub"),
			},
			{
				// Drop the hub behind Terraform's back, then refresh.
				PreConfig: func() {
					for id := range vpnHubState {
						delete(vpnHubState, id)
					}
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_vpn_hub.test", plancheck.ResourceActionCreate),
					},
				},
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
