package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTask(t *testing.T) {
	var mu sync.Mutex
	store := map[string]interface{}{
		"status":      "pending",
		"created_at":  "2024-01-01T00:00:00Z",
		"updated_at":  "2024-01-01T00:00:00Z",
		"task_type":   "reboot",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/tasks":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["name"] = req["name"]
			store["description"] = req["description"]
			store["task_type"] = req["task_type"]
			store["company_id"] = req["company_id"]
			store["payload"] = req["payload"]
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":          1,
				"name":        req["name"],
				"description": req["description"],
				"task_type":   req["task_type"],
				"company_id":  req["company_id"],
				"status":      store["status"],
				"created_at":  "2024-01-01T00:00:00Z",
				"updated_at":  "2024-01-01T00:00:00Z",
			}
			if payload, ok := req["payload"]; ok {
				response["payload"] = payload
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodGet && r.URL.Path == "/tasks/1":
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":          1,
				"name":        store["name"],
				"description": store["description"],
				"task_type":   store["task_type"],
				"company_id":  store["company_id"],
				"status":      store["status"],
				"created_at":  store["created_at"],
				"updated_at":  store["updated_at"],
			}
			if payload, ok := store["payload"]; ok {
				response["payload"] = payload
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodPut && r.URL.Path == "/tasks/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["description"] = req["description"]
			store["updated_at"] = "2024-01-02T00:00:00Z"
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":          1,
				"name":        store["name"],
				"description": store["description"],
				"task_type":   store["task_type"],
				"company_id":  store["company_id"],
				"status":      store["status"],
				"created_at":  store["created_at"],
				"updated_at":  store["updated_at"],
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodDelete && r.URL.Path == "/tasks/1":
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
				Config: testAccTaskConfig(server.URL, "test-task", "Test task description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_task.test", "name", "test-task"),
					resource.TestCheckResourceAttr("rms_task.test", "description", "Test task description"),
					resource.TestCheckResourceAttr("rms_task.test", "task_type", "reboot"),
					resource.TestCheckResourceAttrSet("rms_task.test", "id"),
					resource.TestCheckResourceAttrSet("rms_task.test", "status"),
				),
			},
			{
				Config: testAccTaskConfig(server.URL, "test-task", "Updated description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_task.test", "name", "test-task"),
					resource.TestCheckResourceAttr("rms_task.test", "description", "Updated description"),
				),
			},
		},
	})
}

func testAccTaskConfig(baseURL, name, description string) string {
	return `
provider "rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "rms_task" "test" {
  name        = "` + name + `"
  description = "` + description + `"
  task_type   = "reboot"
  company_id  = 1
  payload     = "{\"command\":\"reboot\"}"
}
`
}
