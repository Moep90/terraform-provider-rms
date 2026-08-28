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

// TestAccPreCheck validates the necessary environment variables are set
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TELTONIKA_RMS_TOKEN"); v == "" {
		t.Fatal("TELTONIKA_RMS_TOKEN must be set for acceptance tests")
	}
}

// ProtoV6ProviderFactories are used to create a test provider
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"teltonika-rms": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
