package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The RMS v3 API returns 405 BAD_REQUEST_METHOD for POST /alerts-configurations,
// so creating this resource is rejected up front. Verified against the live API.
func TestAlertConfigurationResource_CreateRejected(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      testAlertConfigConfig(server.URL, 123, 5, 3, 1, "Test Alert", "Test message", "test@example.com"),
			ExpectError: regexp.MustCompile(`rms_alert_configuration cannot be created`),
		}},
	})
}

func testAlertConfigConfig(baseURL string, deviceID, alertTypeID, alertSubtypeID, action int64, subject, message, email string) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_alert_configuration" "test" {
  device_id          = %d
  alert_type_id      = %d
  alert_subtype_id   = %d
  action             = %d
  subject            = "%s"
  message            = "%s"
  email              = "%s"
}
`, baseURL, deviceID, alertTypeID, alertSubtypeID, action, subject, message, email)
}
