package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func permissionsServer(t *testing.T, payload interface{}) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/permissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Logf("error encoding response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestPermissionsDataSource(t *testing.T) {
	// Returned unsorted, so the sort into name order is observable.
	server := permissionsServer(t, []map[string]interface{}{
		{"id": 5, "name": "view_devices", "description": "View devices"},
		{"id": 2, "name": "create_devices", "description": "Create devices"},
		{"id": 9, "name": "delete_devices", "description": "Delete devices"},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testPermissionsDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_permissions.test", "id", "permissions-data-source"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.#", "3"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.name", "create_devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.id", "2"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.description", "Create devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.1.name", "delete_devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.1.id", "9"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.2.name", "view_devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.2.id", "5"),
			),
		}},
	})
}

// A permission without an id must read as null rather than a fabricated 0.
func TestPermissionsDataSourceMissingID(t *testing.T) {
	server := permissionsServer(t, []map[string]interface{}{
		{"name": "view_devices", "description": "View devices"},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testPermissionsDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.#", "1"),
				resource.TestCheckNoResourceAttr("data.rms_permissions.test", "permissions.0.id"),
			),
		}},
	})
}

// An empty catalogue must yield an empty list, not a null one, so that
// length() and for-expressions over it keep working.
func TestPermissionsDataSourceEmpty(t *testing.T) {
	server := permissionsServer(t, []map[string]interface{}{})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testPermissionsDataSourceConfig(server.URL) + `
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

// The RMS response envelope must be unwrapped transparently.
func TestPermissionsDataSourceEnvelope(t *testing.T) {
	server := permissionsServer(t, map[string]interface{}{
		"success": true,
		"data": []map[string]interface{}{
			{"id": 1, "name": "view_devices", "description": "View devices"},
		},
		"meta": map[string]interface{}{"total": 1},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testPermissionsDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.#", "1"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.name", "view_devices"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.id", "1"),
			),
		}},
	})
}

func testPermissionsDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

data "rms_permissions" "test" {}
`, baseURL)
}
