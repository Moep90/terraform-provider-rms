package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCompany(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/companies":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": req["company_name"],
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": float64(0),
			})

		case r.Method == http.MethodGet && r.URL.Path == "/companies/1":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": "Test Company",
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": float64(0),
			})

		case r.Method == http.MethodPut && r.URL.Path == "/companies/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": req["company_name"],
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": float64(5),
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/companies/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCompanyConfig(server.URL, "Initial Company"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_company.test", "company_name", "Initial Company"),
					resource.TestCheckResourceAttrSet("teltonika-rms_company.test", "id"),
					resource.TestCheckResourceAttrSet("teltonika-rms_company.test", "created_at"),
				),
			},
			{
				Config: testAccCompanyConfig(server.URL, "Updated Company"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_company.test", "company_name", "Updated Company"),
					resource.TestCheckResourceAttrSet("teltonika-rms_company.test", "id"),
				),
			},
			{
				ResourceName:      "teltonika-rms_company.test",
				ImportState:       true,
				ImportStateId:     "1",
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCompanyConfig(baseURL, name string) string {
	return `
provider "teltonika-rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "teltonika-rms_company" "test" {
  company_name = "` + name + `"
}
`
}
