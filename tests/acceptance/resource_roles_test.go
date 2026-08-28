package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRolesDataSource(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			roles := []map[string]interface{}{
				{
					"id":            float64(1),
					"title":         "Admin",
					"description":   "Full admin access",
					"company_id":    float64(123),
					"permission_id": []interface{}{float64(1), float64(2), float64(3)},
				},
				{
					"id":            float64(2),
					"title":         "Device Manager",
					"description":   "Manage devices and tasks",
					"company_id":    float64(123),
					"permission_id": []interface{}{float64(10), float64(11)},
				},
				{
					"id":            float64(3),
					"title":         "Viewer",
					"description":   "Read-only access",
					"company_id":    float64(123),
					"permission_id": []interface{}{float64(20)},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(roles); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testRolesDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_roles.test", "id", "roles-data-source"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.#", "3"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.id", "1"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.title", "Admin"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.description", "Full admin access"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.company_id", "123"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.0.permission_ids.*", "1"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.0.permission_ids.*", "2"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.0.permission_ids.*", "3"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.id", "2"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.title", "Device Manager"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.description", "Manage devices and tasks"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.2.id", "3"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.2.title", "Viewer"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.2.description", "Read-only access"),
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
