package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTag(t *testing.T) {
	// Stateful mock storage
	var mu sync.Mutex
	store := map[string]interface{}{
		"device_count": float64(0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/tags":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["name"] = req["name"]
			store["color"] = req["color"]
			store["company_id"] = req["company_id"]
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"name":         req["name"],
				"color":        req["color"],
				"company_id":   req["company_id"],
				"device_count": store["device_count"],
			})

		case r.Method == http.MethodGet && r.URL.Path == "/tags/1":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"name":         store["name"],
				"color":        store["color"],
				"company_id":   store["company_id"],
				"device_count": store["device_count"],
			})

		case r.Method == http.MethodPut && r.URL.Path == "/tags/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["name"] = req["name"]
			store["color"] = req["color"]
			store["device_count"] = float64(3)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"name":         req["name"],
				"color":        req["color"],
				"company_id":   store["company_id"],
				"device_count": store["device_count"],
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/tags/1":
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
				Config: testAccTagConfig(server.URL, "tag-a", "#00ff00"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_tag.test", "name", "tag-a"),
					resource.TestCheckResourceAttr("teltonika-rms_tag.test", "color", "#00ff00"),
					resource.TestCheckResourceAttrSet("teltonika-rms_tag.test", "id"),
				),
			},
			{
				Config: testAccTagConfig(server.URL, "tag-b", "#0000ff"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_tag.test", "name", "tag-b"),
					resource.TestCheckResourceAttr("teltonika-rms_tag.test", "color", "#0000ff"),
					resource.TestCheckResourceAttrSet("teltonika-rms_tag.test", "id"),
				),
			},
		},
	})
}

func testAccTagConfig(baseURL, name, color string) string {
	return `
provider "teltonika-rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "teltonika-rms_tag" "test" {
  name       = "` + name + `"
  color      = "` + color + `"
  company_id = 1
}
`
}
