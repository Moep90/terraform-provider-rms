package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDevice(t *testing.T) {
	var mu sync.Mutex
	store := map[string]interface{}{
		"status":             "not_activated",
		"firmware":           "2.11",
		"created_at":         "2024-01-01T00:00:00Z",
		"auto_credit_enable": true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["name"] = req["name"]
			store["device_series"] = req["device_series"]
			store["serial"] = req["serial"]
			store["company_id"] = req["company_id"]
			if mac, ok := req["mac"].(string); ok {
				store["mac"] = mac
			}
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":                 1,
				"name":               req["name"],
				"device_series":      req["device_series"],
				"serial":             req["serial"],
				"company_id":         req["company_id"],
				"status":             store["status"],
				"firmware":           store["firmware"],
				"created_at":         store["created_at"],
				"auto_credit_enable": store["auto_credit_enable"],
			}
			if mac, ok := store["mac"]; ok {
				response["mac"] = mac
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodGet && r.URL.Path == "/devices/1":
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":                 1,
				"name":               store["name"],
				"device_series":      store["device_series"],
				"serial":             store["serial"],
				"company_id":         store["company_id"],
				"status":             store["status"],
				"firmware":           store["firmware"],
				"created_at":         store["created_at"],
				"auto_credit_enable": store["auto_credit_enable"],
			}
			if mac, ok := store["mac"]; ok {
				response["mac"] = mac
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodPut && r.URL.Path == "/devices/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			store["name"] = req["name"]
			if s, ok := req["status"].(string); ok {
				store["status"] = s
			}
			if f, ok := req["firmware"].(string); ok {
				store["firmware"] = f
			}
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":                 1,
				"name":               req["name"],
				"device_series":      store["device_series"],
				"serial":             store["serial"],
				"company_id":         store["company_id"],
				"status":             store["status"],
				"firmware":           store["firmware"],
				"created_at":         store["created_at"],
				"auto_credit_enable": store["auto_credit_enable"],
			}
			_ = json.NewEncoder(w).Encode(response)

		case r.Method == http.MethodDelete && r.URL.Path == "/devices/1":
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
				Config: testAccDeviceConfig(server.URL, "dev-a", "rut"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device.test", "name", "dev-a"),
					resource.TestCheckResourceAttr("rms_device.test", "device_series", "rut"),
					resource.TestCheckResourceAttrSet("rms_device.test", "id"),
					resource.TestCheckResourceAttrSet("rms_device.test", "status"),
					resource.TestCheckResourceAttrSet("rms_device.test", "firmware"),
				),
			},
			{
				Config: testAccDeviceConfig(server.URL, "dev-b", "rut"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device.test", "name", "dev-b"),
					resource.TestCheckResourceAttr("rms_device.test", "device_series", "rut"),
					resource.TestCheckResourceAttrSet("rms_device.test", "id"),
					resource.TestCheckResourceAttrSet("rms_device.test", "status"),
					resource.TestCheckResourceAttrSet("rms_device.test", "firmware"),
				),
			},
		},
	})
}

func testAccDeviceConfig(baseURL, name, series string) string {
	return `
provider "rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "rms_device" "test" {
  name            = "` + name + `"
  device_series   = "` + series + `"
  serial          = "0123456789"
  company_id      = 1
  auto_credit_enable = true
}
`
}
