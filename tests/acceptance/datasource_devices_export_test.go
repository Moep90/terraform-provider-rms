package acceptance

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDevicesExportDataSource_Read(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/devices/export/csv", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)

		writer := csv.NewWriter(w)
		if err := writer.Write([]string{"id", "name", "serial", "device_series", "status"}); err != nil {
			t.Logf("error writing CSV header: %v", err)
		}
		if err := writer.Write([]string{"123", "Device 1", "SN001", "rut", "online"}); err != nil {
			t.Logf("error writing CSV row: %v", err)
		}
		if err := writer.Write([]string{"456", "Device 2", "SN002", "trb", "offline"}); err != nil {
			t.Logf("error writing CSV row: %v", err)
		}
		writer.Flush()
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDevicesExportConfig(server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.rms_devices_export.test", "csv_data", "id,name,serial,device_series,status\n123,Device 1,SN001,rut,online\n456,Device 2,SN002,trb,offline\n"),
				),
			},
		},
	})
}

func testDevicesExportConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

data "rms_devices_export" "test" {}
`, baseURL)
}
