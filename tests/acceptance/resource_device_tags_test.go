package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDeviceTagsResource_Assign(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	deviceTagState := map[int][]int{123: {10, 20}}

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":                 123,
				"name":               "Test Device",
				"device_series":      "rut",
				"serial":             "SN123",
				"company_id":         float64(1),
				"mac":                "00:11:22:33:44:55",
				"status":             "online",
				"firmware":           "v1.0",
				"created_at":         "2024-01-01T00:00:00Z",
				"monitoring_enable":  false,
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 9 {
			var deviceID int
			if _, err := fmt.Sscanf(path[9:], "%d", &deviceID); err != nil {
				t.Logf("error parsing device ID: %v", err)
			}

			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				resp := map[string]interface{}{
					"id":              deviceID,
					"name":            "Test Device",
					"device_series":   "rut",
					"status":          "online",
					"firmware":        "v1.0",
					"created_at":      "2024-01-01T00:00:00Z",
					"monitoring_enable": false,
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					t.Logf("error encoding response: %v", err)
				}
				return
			}

			if r.Method == http.MethodPut {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Logf("error decoding request: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				resp := map[string]interface{}{
					"id":              deviceID,
					"name":            "Test Device",
					"device_series":   "rut",
					"status":          "online",
					"firmware":        "v1.0",
					"created_at":      "2024-01-01T00:00:00Z",
					"monitoring_enable": false,
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					t.Logf("error encoding response: %v", err)
				}
				return
			}

			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/devices/123/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			tagIDs := deviceTagState[123]
			tags := make([]interface{}, len(tagIDs))
			for i, id := range tagIDs {
				tags[i] = map[string]interface{}{
					"id":   id,
					"name": fmt.Sprintf("Tag %d", id),
				}
			}
			if err := json.NewEncoder(w).Encode(tags); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}
			if tagIdsRaw, ok := req["tag_ids"]; ok {
				if tagIds, ok := tagIdsRaw.([]interface{}); ok {
					ids := []int{}
					for _, tid := range tagIds {
						if f, ok := tid.(float64); ok {
							ids = append(ids, int(f))
						}
					}
					deviceTagState[123] = ids
				}
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"success": true}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDeviceTagsConfig(server.URL, []int{10, 20}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device_tags.assignment", "device_id", "123"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "20"),
				),
			},
			{
				Config: testDeviceTagsConfig(server.URL, []int{10, 20, 30}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device_tags.assignment", "device_id", "123"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "20"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "30"),
				),
			},
			{
				Config: testDeviceTagsConfig(server.URL, []int{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device_tags.assignment", "device_id", "123"),
				),
			},
		},
	})
}

func testDeviceTagsConfig(baseURL string, tagIDs []int) string {
	tagIDsStr := "["
	for i, id := range tagIDs {
		if i > 0 {
			tagIDsStr += ", "
		}
		tagIDsStr += fmt.Sprintf("%d", id)
	}
	tagIDsStr += "]"

	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_device" "test" {
  name              = "Test Device"
  device_series     = "rut"
  serial            = "SN123"
  company_id        = 1
  mac               = "00:11:22:33:44:55"
  monitoring_enable = false
}

resource "rms_device_tags" "assignment" {
  device_id = rms_device.test.id
  tag_ids   = %s
}
`, baseURL, tagIDsStr)
}
