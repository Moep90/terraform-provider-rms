package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// rolesServer mirrors the live RMS shape: the /roles list is wrapped in an
// envelope and carries permissions_count but no permission IDs, so the IDs are
// served per role from /roles/{id}/permissions.
func rolesServer(t *testing.T, roles []map[string]interface{}, perms map[int][]map[string]interface{}) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    roles,
			"meta":    map[string]interface{}{"total": len(roles)},
		}); err != nil {
			t.Logf("error encoding response: %v", err)
		}
	})
	for id, list := range perms {
		list := list
		mux.HandleFunc(fmt.Sprintf("/roles/%d/permissions", id), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    list,
			}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
		})
	}

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestRolesDataSource(t *testing.T) {
	// company_id is null on RMS built-in roles, and the roles arrive out of
	// id order.
	server := rolesServer(t,
		[]map[string]interface{}{
			{"id": 6, "title": "Advanced guest", "name": "readonly_admin", "description": "Read only", "company_id": nil, "permissions_count": 2},
			{"id": 2, "title": "Administrator", "name": "admin", "description": "Everything", "company_id": 123, "permissions_count": 1},
		},
		map[int][]map[string]interface{}{
			6: {
				{"id": 28, "name": "view_pending_device_actions", "title": "View pending", "category": "Device actions"},
				{"id": 5, "name": "create_devices", "title": "Create devices", "category": "Devices"},
			},
			2: {{"id": 9, "name": "manage_all", "title": "Manage all", "category": "Admin"}},
		})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testRolesDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_roles.test", "id", "roles-data-source"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.#", "2"),
				// sorted by id
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.id", "2"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.title", "Administrator"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.name", "admin"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.company_id", "123"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.0.permission_ids.*", "9"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.id", "6"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.name", "readonly_admin"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.permission_ids.#", "2"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.1.permission_ids.*", "28"),
			),
		}},
	})
}

// company_id is null on every RMS built-in role; it must not read as 0.
func TestRolesDataSourceNullCompanyID(t *testing.T) {
	server := rolesServer(t,
		[]map[string]interface{}{{"id": 6, "title": "Advanced guest", "name": "readonly_admin", "company_id": nil}},
		map[int][]map[string]interface{}{6: {}})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testRolesDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr("data.rms_roles.test", "roles.0.company_id"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.permission_ids.#", "0"),
			),
		}},
	})
}

// A malformed row must name the role it came from.
func TestRolesDataSourceNullTitle(t *testing.T) {
	server := rolesServer(t,
		[]map[string]interface{}{{"id": 2, "title": nil}},
		map[int][]map[string]interface{}{2: {}})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      testRolesDataSourceConfig(server.URL),
			ExpectError: regexp.MustCompile(`Role 2 has no string title`),
		}},
	})
}

// An empty role list must yield an empty list, not a null one.
func TestRolesDataSourceEmpty(t *testing.T) {
	server := rolesServer(t, []map[string]interface{}{}, nil)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testRolesDataSourceConfig(server.URL) + `
output "count" {
  value = length(data.rms_roles.test.roles)
}
`,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.#", "0"),
				resource.TestCheckOutput("count", "0"),
			),
		}},
	})
}

func testRolesDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

data "rms_roles" "test" {}
`, baseURL)
}
