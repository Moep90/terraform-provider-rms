package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDeviceEsimBootstrapDataSource_Read(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/devices/123/check/esim-bootstrap", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"esim_bootstrap": "enabled",
			"status":         "success",
			"message":        "E-SIM is active",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Logf("error encoding response: %v", err)
		}
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDeviceEsimBootstrapConfig(server.URL, 123),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.rms_device_esim_bootstrap.test", "device_id", "123"),
					resource.TestCheckResourceAttr("data.rms_device_esim_bootstrap.test", "esim_bootstrap", "enabled"),
					resource.TestCheckResourceAttr("data.rms_device_esim_bootstrap.test", "status", "success"),
					resource.TestCheckResourceAttr("data.rms_device_esim_bootstrap.test", "message", "E-SIM is active"),
				),
			},
		},
	})
}

func TestDeviceEsimBootstrapDataSource_ReadDisabled(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/devices/456/check/esim-bootstrap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"esim_bootstrap": "disabled",
			"status":         "success",
			"message":        "No E-SIM detected",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Logf("error encoding response: %v", err)
		}
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDeviceEsimBootstrapConfig(server.URL, 456),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.rms_device_esim_bootstrap.test", "device_id", "456"),
					resource.TestCheckResourceAttr("data.rms_device_esim_bootstrap.test", "esim_bootstrap", "disabled"),
				),
			},
		},
	})
}

func testDeviceEsimBootstrapConfig(baseURL string, deviceID int) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

data "rms_device_esim_bootstrap" "test" {
  device_id = %d
}
`, baseURL, deviceID)
}
