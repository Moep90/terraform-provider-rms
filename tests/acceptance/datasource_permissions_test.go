package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// permissionsServer mirrors the live RMS shape: every response is wrapped in
// {"success":true,"data":[...]} and permissions are reachable only through a
// role at /roles/{id}/permissions. There is no global /permissions endpoint.
func permissionsServer(t *testing.T, roleID int, permissions []map[string]interface{}) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/roles/%d/permissions", roleID), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    permissions,
		}); err != nil {
			t.Logf("error encoding response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestPermissionsDataSource(t *testing.T) {
	// Returned unsorted so the sort into name order is observable.
	server := permissionsServer(t, 2, []map[string]interface{}{
		{"id": 28, "name": "view_pending_device_actions", "title": "View pending device actions", "description": "", "category": "Device actions"},
		{"id": 5, "name": "create_devices", "title": "Create devices", "description": "Add devices", "category": "Devices"},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testPermissionsDataSourceConfig(server.URL, 2),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_permissions.test", "id", "role-2-permissions"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "role_id", "2"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.#", "2"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.name", "create_devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.id", "5"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.title", "Create devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.category", "Devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.1.name", "view_pending_device_actions"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.1.id", "28"),
			),
		}},
	})
}

// A role with no permissions must yield an empty list, not a null one, so
// length() and for-expressions over it keep working.
func TestPermissionsDataSourceEmpty(t *testing.T) {
	server := permissionsServer(t, 2, []map[string]interface{}{})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testPermissionsDataSourceConfig(server.URL, 2) + `
output "count" {
  value = length(data.rms_permissions.test.permissions)
}
`,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.#", "0"),
				resource.TestCheckOutput("count", "0"),
			),
		}},
	})
}

func testPermissionsDataSourceConfig(baseURL string, roleID int) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

data "rms_permissions" "test" {
  role_id = %d
}
`, baseURL, roleID)
}
