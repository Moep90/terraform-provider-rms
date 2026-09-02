package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDeviceResource_Monitoring(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Track monitoring state
	monitoringState := map[int]bool{123: false}

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":                 123,
				"name":               req["name"],
				"device_series":      req["device_series"],
				"serial":             req["serial"],
				"company_id":         req["company_id"],
				"auto_credit_enable": req["auto_credit_enable"],
				"status":             "active",
				"firmware":           "v1.0",
				"created_at":         "2024-01-01T00:00:00Z",
				"mac":                nil,
				"imei":               nil,
				"monitoring_enable":  false,
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		// RMS deletes devices through the collection, selecting them by id in
		// the request body.
		if r.Method == http.MethodDelete {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("DELETE /devices must carry a JSON body: %v", err)
			}
			if _, ok := req["device_id"].([]interface{}); !ok {
				t.Errorf("DELETE /devices body has no device_id list: %v", req)
			}
			w.WriteHeader(http.StatusOK)
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
					"id":                deviceID,
					"name":              "Test Device",
					"device_series":     "rut",
					"serial":            "SN123",
					"company_id":        float64(1),
					"mac":               nil,
					"imei":              nil,
					"status":            "active",
					"firmware":          "v1.0",
					"created_at":        "2024-01-01T00:00:00Z",
					"monitoring_enable": monitoringState[deviceID],
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
				// Handle monitoring_enable from device update
				if monitoringEnable, ok := req["monitoring_enable"].(bool); ok {
					monitoringState[deviceID] = monitoringEnable
				}
				w.Header().Set("Content-Type", "application/json")
				resp := map[string]interface{}{
					"id":                deviceID,
					"name":              "Test Device",
					"status":            "active",
					"firmware":          "v1.0",
					"created_at":        "2024-01-01T00:00:00Z",
					"mac":               nil,
					"imei":              nil,
					"monitoring_enable": monitoringState[deviceID],
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					t.Logf("error encoding response: %v", err)
				}
				return
			}

		}

		w.WriteHeader(http.StatusNotFound)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDeviceMonitoringConfig(server.URL, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device.monitoring_test", "monitoring_enable", "false"),
				),
			},
			{
				Config: testDeviceMonitoringConfig(server.URL, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device.monitoring_test", "monitoring_enable", "true"),
				),
			},
		},
	})
}

func testDeviceMonitoringConfig(baseURL string, monitoring bool) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_device" "monitoring_test" {
  name             = "Test Device"
  device_series    = "rut"
  serial           = "SN123"
  company_id       = 1
  monitoring_enable = %t
}
`, baseURL, monitoring)
}
