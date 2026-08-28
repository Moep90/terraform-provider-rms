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

func TestRoleResource_CRUD(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	roleState := make(map[int]map[string]interface{})
	nextID := 1

	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}

			currentID := nextID
			nextID++

			title := ""
			if t, ok := req["title"].(string); ok {
				title = t
			}
			description := ""
			if d, ok := req["description"].(string); ok {
				description = d
			}
			var permIDs []interface{}
			if p, ok := req["permission_id"].([]interface{}); ok {
				permIDs = p
			}

			roleState[currentID] = map[string]interface{}{
				"id":            float64(currentID),
				"title":         title,
				"description":   description,
				"company_id":    float64(1),
				"permission_id": permIDs,
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": float64(currentID)}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/roles/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		parts := strings.Split(path, "/")

		// /roles/{id}/permissions - the live API serves permission IDs only
		// here; a role read itself carries permissions_count.
		if len(parts) >= 4 && parts[len(parts)-1] == "permissions" {
			id, err := strconv.Atoi(parts[len(parts)-2])
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			cfg, ok := roleState[id]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			perms := []map[string]interface{}{}
			if ids, ok := cfg["permission_id"].([]interface{}); ok {
				for _, raw := range ids {
					var n float64
					switch v := raw.(type) {
					case float64:
						n = v
					case int:
						n = float64(v)
					default:
						continue
					}
					perms = append(perms, map[string]interface{}{
						"id":   n,
						"name": fmt.Sprintf("perm_%d", int(n)),
					})
				}
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    perms,
			}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		if len(parts) >= 3 {
			id, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if r.Method == http.MethodGet {
				if cfg, ok := roleState[id]; ok {
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
				if cfg, ok := roleState[id]; ok {
					if title, ok := req["title"].(string); ok {
						cfg["title"] = title
					}
					if description, ok := req["description"].(string); ok {
						cfg["description"] = description
					}
					if permIDs, ok := req["permission_id"].([]interface{}); ok {
						cfg["permission_id"] = permIDs
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

			if r.Method == http.MethodDelete {
				delete(roleState, id)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRoleConfig(server.URL, "Admin Role", "Full admin access", 1, []int{10, 20, 30}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_role.test", plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_role.test", "title", "Admin Role"),
					resource.TestCheckResourceAttr("rms_role.test", "description", "Full admin access"),
					resource.TestCheckResourceAttr("rms_role.test", "company_id", "1"),
					resource.TestCheckTypeSetElemAttr("rms_role.test", "permission_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_role.test", "permission_ids.*", "20"),
					resource.TestCheckTypeSetElemAttr("rms_role.test", "permission_ids.*", "30"),
				),
			},
			{
				// title has RequiresReplace, so this step is a replacement, not
				// an update. Asserting it pins the plan modifier in place.
				Config: testRoleConfig(server.URL, "Admin Role Updated", "Updated admin access", 1, []int{10, 20}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_role.test", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_role.test", "title", "Admin Role Updated"),
					resource.TestCheckResourceAttr("rms_role.test", "description", "Updated admin access"),
					resource.TestCheckTypeSetElemAttr("rms_role.test", "permission_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_role.test", "permission_ids.*", "20"),
				),
			},
			{
				// Only the mutable fields change, so this must be an in-place
				// update. A dropped plan modifier turning it into a replacement
				// is invisible to a state check.
				Config: testRoleConfig(server.URL, "Admin Role Updated", "Updated again", 1, []int{10}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_role.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_role.test", "description", "Updated again"),
					resource.TestCheckResourceAttr("rms_role.test", "permission_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("rms_role.test", "permission_ids.*", "10"),
				),
			},
		},
	})
}

func testRoleConfig(baseURL string, title, description string, companyID int64, permissionIDs []int) string {
	permStr := "["
	for i, id := range permissionIDs {
		if i > 0 {
			permStr += ", "
		}
		permStr += fmt.Sprintf("%d", id)
	}
	permStr += "]"

	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_role" "test" {
  title           = "%s"
  description     = "%s"
  company_id      = %d
  permission_ids  = %s
}
`, baseURL, title, description, companyID, permStr)
}
