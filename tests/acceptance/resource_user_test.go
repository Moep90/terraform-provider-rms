package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUser(t *testing.T) {
	var mu sync.Mutex
	store := map[string]interface{}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/users":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["username"] = req["username"]
			store["email"] = req["email"]
			store["role"] = req["role"]
			store["company_id"] = req["company_id"]
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
				"username":   store["username"],
				"email":      store["email"],
				"role":       store["role"],
				"company_id": store["company_id"],
			})

		case r.Method == http.MethodPut && r.URL.Path == "/users/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["username"] = req["username"]
			store["email"] = req["email"]
			store["role"] = req["role"]
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1,
				"username":   store["username"],
				"email":      store["email"],
				"role":       store["role"],
				"company_id": store["company_id"],
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
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(server.URL, "user-a", "a@example.com", "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_user.test", "username", "user-a"),
					resource.TestCheckResourceAttr("rms_user.test", "email", "a@example.com"),
					resource.TestCheckResourceAttr("rms_user.test", "role", "viewer"),
					resource.TestCheckResourceAttrSet("rms_user.test", "id"),
				),
			},
			{
				Config: testAccUserConfig(server.URL, "user-b", "b@example.com", "admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_user.test", "username", "user-b"),
					resource.TestCheckResourceAttr("rms_user.test", "email", "b@example.com"),
					resource.TestCheckResourceAttr("rms_user.test", "role", "admin"),
					resource.TestCheckResourceAttrSet("rms_user.test", "id"),
				),
			},
		},
	})
}

func testAccUserConfig(baseURL, username, email, role string) string {
	return `
provider "rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "rms_user" "test" {
  username   = "` + username + `"
  email      = "` + email + `"
  role       = "` + role + `"
  company_id = 1
}
`
}
