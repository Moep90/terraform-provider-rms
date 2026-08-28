package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTaskGroup(t *testing.T) {
	var mu sync.Mutex
	store := map[string]interface{}{
		"status":     "active",
		"task_count": float64(0),
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/task-groups":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["name"] = req["name"]
			store["description"] = req["description"]
			store["company_id"] = req["company_id"]
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":          1,
				"name":        req["name"],
				"description": req["description"],
				"company_id":  req["company_id"],
				"status":      store["status"],
				"task_count":  store["task_count"],
				"created_at":  "2024-01-01T00:00:00Z",
				"updated_at":  "2024-01-01T00:00:00Z",
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodGet && r.URL.Path == "/task-groups/1":
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":          1,
				"name":        store["name"],
				"description": store["description"],
				"company_id":  store["company_id"],
				"status":      store["status"],
				"task_count":  store["task_count"],
				"created_at":  store["created_at"],
				"updated_at":  store["updated_at"],
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodPut && r.URL.Path == "/task-groups/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["name"] = req["name"]
			store["description"] = req["description"]
			store["updated_at"] = "2024-01-02T00:00:00Z"
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":          1,
				"name":        store["name"],
				"description": store["description"],
				"company_id":  store["company_id"],
				"status":      store["status"],
				"task_count":  store["task_count"],
				"created_at":  store["created_at"],
				"updated_at":  store["updated_at"],
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodDelete && r.URL.Path == "/task-groups/1":
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
				Config: testAccTaskGroupConfig(server.URL, "test-task-group", "Test task group description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_task_group.test", "name", "test-task-group"),
					resource.TestCheckResourceAttr("teltonika-rms_task_group.test", "description", "Test task group description"),
					resource.TestCheckResourceAttrSet("teltonika-rms_task_group.test", "id"),
					resource.TestCheckResourceAttrSet("teltonika-rms_task_group.test", "status"),
					resource.TestCheckResourceAttrSet("teltonika-rms_task_group.test", "task_count"),
				),
			},
			{
				Config: testAccTaskGroupConfig(server.URL, "test-task-group-updated", "Updated description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_task_group.test", "name", "test-task-group-updated"),
					resource.TestCheckResourceAttr("teltonika-rms_task_group.test", "description", "Updated description"),
				),
			},
		},
	})
}

func testAccTaskGroupConfig(baseURL, name, description string) string {
	return `
provider "teltonika-rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "teltonika-rms_task_group" "test" {
  name        = "` + name + `"
  description = "` + description + `"
  company_id  = 1
}
`
}
