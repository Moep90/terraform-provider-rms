package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/users":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1,
				"username":   req["username"],
				"email":      req["email"],
				"role":       req["role"],
				"company_id": req["company_id"],
			})

		case r.Method == http.MethodGet && r.URL.Path == "/users/1":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1,
				"username":   "testuser",
				"email":      "test@example.com",
				"role":       "admin",
				"company_id": float64(1),
			})

		case r.Method == http.MethodPut && r.URL.Path == "/users/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1,
				"username":   "testuser",
				"email":      req["email"],
				"role":       req["role"],
				"company_id": float64(1),
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/users/1":
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
				Config: testAccUserConfig(server.URL, "user-a", "a@example.com", "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_user.test", "username", "user-a"),
					resource.TestCheckResourceAttr("teltonika-rms_user.test", "email", "a@example.com"),
					resource.TestCheckResourceAttr("teltonika-rms_user.test", "role", "viewer"),
					resource.TestCheckResourceAttrSet("teltonika-rms_user.test", "id"),
				),
			},
			{
				Config: testAccUserConfig(server.URL, "user-b", "b@example.com", "admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_user.test", "username", "user-b"),
					resource.TestCheckResourceAttr("teltonika-rms_user.test", "email", "b@example.com"),
					resource.TestCheckResourceAttr("teltonika-rms_user.test", "role", "admin"),
					resource.TestCheckResourceAttrSet("teltonika-rms_user.test", "id"),
				),
			},
		},
	})
}

func testAccUserConfig(baseURL, username, email, role string) string {
	return `
provider "teltonika-rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "teltonika-rms_user" "test" {
  username   = "` + username + `"
  email      = "` + email + `"
  role       = "` + role + `"
  company_id = 1
}
`
}
