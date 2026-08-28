package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/provider"
)

const (
	// Provider address for testing
	ProviderAddr = "registry.terraform.io/teltonika-rms/teltonika-rms"
)

// ProtoV6ProviderFactories are used to create a test provider
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"rms": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// isRealAPITest returns true if we should test against the real API
func isRealAPITest() bool {
	// E2E tests require both RMS_ADMIN_TOKEN and explicit RUN_E2E_TESTS flag
	// They are skipped by default to avoid failures when API is unavailable
	return os.Getenv("RMS_ADMIN_TOKEN") != "" && os.Getenv("RUN_E2E_TESTS") == "true"
}

// testAccPreCheck validates that required environment variables are set
func testAccPreCheck(t *testing.T) {
	if !isRealAPITest() {
		if v := os.Getenv("TELTONIKA_RMS_TOKEN"); v == "" {
			// Skip only if TELTONIKA_RMS_FORCE_TEST is not set
			if os.Getenv("TELTONIKA_RMS_FORCE_TEST") == "" {
				t.Skip("TELTONIKA_RMS_TOKEN must be set for acceptance tests (or set TELTONIKA_RMS_FORCE_TEST=1 to use mocks)")
			}
		}
	}
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
