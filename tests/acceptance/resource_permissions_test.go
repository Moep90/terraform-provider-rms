package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPermissionsDataSource(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/permissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			permissions := []map[string]interface{}{
				{"name": "view_pending_device_actions", "description": "View pending device actions"},
				{"name": "execute_device_actions", "description": "Execute device actions"},
				{"name": "manage_device_passwords", "description": "Manage device passwords"},
				{"name": "update_device_credit_status", "description": "Update device credit status"},
				{"name": "view_devices", "description": "View devices"},
				{"name": "create_devices", "description": "Create devices"},
				{"name": "update_devices", "description": "Update devices"},
				{"name": "delete_devices", "description": "Delete devices"},
				{"name": "view_roles", "description": "View user roles"},
				{"name": "create_roles", "description": "Create roles"},
				{"name": "update_roles", "description": "Update roles"},
				{"name": "delete_roles", "description": "Delete roles"},
				{"name": "view_permissions", "description": "View permissions"},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(permissions); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testPermissionsDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_permissions.test", "id", "permissions-data-source"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.#", "13"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.name", "view_pending_device_actions"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.0.description", "View pending device actions"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.1.name", "execute_device_actions"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.1.description", "Execute device actions"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.2.name", "manage_device_passwords"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.2.description", "Manage device passwords"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.12.name", "view_permissions"),
				resource.TestCheckResourceAttr("data.rms_permissions.test", "permissions.12.description", "View permissions"),
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
