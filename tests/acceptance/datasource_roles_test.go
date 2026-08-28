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

func rolesServer(t *testing.T, payload interface{}) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
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

func TestRolesDataSource(t *testing.T) {
	// company_id is echoed as the single-element array that role writes send,
	// and the roles arrive out of id order.
	server := rolesServer(t, []map[string]interface{}{
		{
			"id":            3,
			"title":         "Viewer",
			"description":   "Read-only access",
			"company_id":    []interface{}{123},
			"permission_id": []interface{}{20},
		},
		{
			"id":            1,
			"title":         "Admin",
			"description":   "Full admin access",
			"company_id":    123,
			"permission_id": []interface{}{1, 2, 3},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testRolesDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_roles.test", "id", "roles-data-source"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.#", "2"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.id", "1"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.title", "Admin"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.description", "Full admin access"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.company_id", "123"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.0.permission_ids.*", "1"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.0.permission_ids.*", "3"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.id", "3"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.title", "Viewer"),
				// company_id delivered as an array must resolve, not read as 0.
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.1.company_id", "123"),
			),
		}},
	})
}

// A role without permission_id must expose an empty set, not a null one:
// rms_role.permission_ids is required and cannot accept null.
func TestRolesDataSourceMissingPermissions(t *testing.T) {
	server := rolesServer(t, []map[string]interface{}{
		{"id": 1, "title": "Admin", "company_id": 5},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testRolesDataSourceConfig(server.URL) + `
output "perm_count" {
  value = length(data.rms_roles.test.roles[0].permission_ids)
}
`,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.permission_ids.#", "0"),
				resource.TestCheckOutput("perm_count", "0"),
			),
		}},
	})
}

// Non-numeric permission ids must fail loudly. Dropping them would create a
// role with fewer privileges than configured while reporting success.
func TestRolesDataSourceNonNumericPermissionID(t *testing.T) {
	server := rolesServer(t, []map[string]interface{}{
		{"id": 1, "title": "Admin", "company_id": 5, "permission_id": []interface{}{"1", "2"}},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      testRolesDataSourceConfig(server.URL),
			ExpectError: regexp.MustCompile(`permission_id\[0\] is string \(1\), want a number`),
		}},
	})
}

// A malformed row must name the role it came from.
func TestRolesDataSourceNullTitle(t *testing.T) {
	server := rolesServer(t, []map[string]interface{}{
		{"id": 1, "title": "Ok"},
		{"id": 2, "title": nil},
	})

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
	server := rolesServer(t, []map[string]interface{}{})

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

// The RMS response envelope must be unwrapped transparently.
func TestRolesDataSourceEnvelope(t *testing.T) {
	server := rolesServer(t, map[string]interface{}{
		"success": true,
		"data": []map[string]interface{}{
			{"id": 1, "title": "Admin", "company_id": 5, "permission_id": []interface{}{7}},
		},
		"meta": map[string]interface{}{"total": 1},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testRolesDataSourceConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.#", "1"),
				resource.TestCheckResourceAttr("data.rms_roles.test", "roles.0.title", "Admin"),
				resource.TestCheckTypeSetElemAttr("data.rms_roles.test", "roles.0.permission_ids.*", "7"),
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
