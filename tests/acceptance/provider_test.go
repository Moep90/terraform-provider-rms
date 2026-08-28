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
	"teltonika-rms": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck validates that required environment variables are set
// For mocked acceptance tests, this is optional and will skip only if explicitly required
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TELTONIKA_RMS_TOKEN"); v == "" {
		// Skip only if TELTONIKA_RMS_FORCE_TEST is not set
		if os.Getenv("TELTONIKA_RMS_FORCE_TEST") == "" {
			t.Skip("TELTONIKA_RMS_TOKEN must be set for acceptance tests (or set TELTONIKA_RMS_FORCE_TEST=1 to use mocks)")
		}
	}
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
