package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":                 1,
				"name":               req["name"],
				"device_series":      req["device_series"],
				"serial":             req["serial"],
				"company_id":         req["company_id"],
				"status":             "not_activated",
				"firmware":           "2.11",
				"created_at":         "2024-01-01T00:00:00Z",
				"auto_credit_enable": true,
			})

		case r.Method == http.MethodGet && r.URL.Path == "/devices/1":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":                 1,
				"name":               "Test Device",
				"device_series":      "rut",
				"serial":             "0123456789",
				"company_id":         float64(1),
				"mac":                "00:11:22:33:44:55",
				"status":             "online",
				"firmware":           "2.12",
				"created_at":         "2024-01-01T00:00:00Z",
				"auto_credit_enable": true,
			})

		case r.Method == http.MethodPut && r.URL.Path == "/devices/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":                 1,
				"name":               req["name"],
				"device_series":      "rut",
				"serial":             "0123456789",
				"company_id":         float64(1),
				"status":             "online",
				"firmware":           "2.12",
				"created_at":         "2024-01-01T00:00:00Z",
				"auto_credit_enable": true,
			})

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
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfig(server.URL, "dev-a", "rut"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_device.test", "name", "dev-a"),
					resource.TestCheckResourceAttr("teltonika-rms_device.test", "device_series", "rut"),
					resource.TestCheckResourceAttrSet("teltonika-rms_device.test", "id"),
					resource.TestCheckResourceAttrSet("teltonika-rms_device.test", "status"),
					resource.TestCheckResourceAttrSet("teltonika-rms_device.test", "firmware"),
				),
			},
			{
				Config: testAccDeviceConfig(server.URL, "dev-b", "rut"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("teltonika-rms_device.test", "name", "dev-b"),
					resource.TestCheckResourceAttrSet("teltonika-rms_device.test", "id"),
					resource.TestCheckResourceAttrSet("teltonika-rms_device.test", "status"),
					resource.TestCheckResourceAttrSet("teltonika-rms_device.test", "firmware"),
				),
			},
		},
	})
}

func testAccDeviceConfig(baseURL, name, series string) string {
	return `
provider "teltonika-rms" {
  token     = "test-token"
  base_url  = "` + baseURL + `"
}

resource "teltonika-rms_device" "test" {
  name            = "` + name + `"
  device_series   = "` + series + `"
  serial          = "0123456789"
  company_id      = 1
  auto_credit_enable = true
}
`
}
