package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// taskServer serves the task operations RMS defines: GET on /devices/tasks and
// PUT and DELETE on /devices/tasks/{id}. There is no create and no read-by-id,
// so POST /devices/tasks and GET /devices/tasks/{id} 404 the way RMS does.
func taskServer(t *testing.T) *httptest.Server {
	t.Helper()

	var mu sync.Mutex
	store := map[string]interface{}{
		"id":          float64(1),
		"name":        "test-task",
		"description": "Test task description",
		"type":        "reboot",
		"status":      "pending",
		"company_id":  float64(1),
		"payload":     `{"command":"reboot"}`,
		"created_at":  "2024-01-01T00:00:00Z",
		"updated_at":  "2024-01-01T00:00:00Z",
	}
	present := true

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/devices/tasks":
			items := []interface{}{}
			if present {
				items = append(items, store)
			}
			writeRMSList(t, w, items)

		case r.Method == http.MethodPut && r.URL.Path == "/devices/tasks/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			if desc, ok := req["description"]; ok {
				store["description"] = desc
			}
			store["updated_at"] = "2024-01-02T00:00:00Z"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": store})

		case r.Method == http.MethodDelete && r.URL.Path == "/devices/tasks/1":
			present = false
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))

		default:
			// Notably POST /devices/tasks lands here: RMS does not define it.
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return server
}

// TestAccTask_CreateRejected pins the deliberate apply-time failure: RMS
// exposes no create operation for individual tasks, so the provider must say so
// instead of issuing a request that cannot work.
func TestAccTask_CreateRejected(t *testing.T) {
	server := taskServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccTaskConfig(server.URL, "test-task", "Test task description"),
				ExpectError: regexp.MustCompile(`rms_task cannot be created`),
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
