package acceptance

import (
	"os"
	"testing"

	"github.com/Moep90/terraform-provider-rms/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// Provider address for testing
	ProviderAddr = "registry.terraform.io/moep90/rms"
)

// ProtoV6ProviderFactories are used to create a test provider
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"rms": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// isRealAPITest reports whether a credential for the real API is available.
// The token variables are the same ones the provider itself reads, so an
// environment configured to run the provider also runs these tests.
//
// TF_ACC remains the gate that decides whether tests touching real
// infrastructure execute at all: terraform-plugin-testing skips every
// resource.Test without it, so a bare `go test ./...` never reaches RMS.
func isRealAPITest() bool {
	return e2eToken() != ""
}

// e2eToken resolves the API token from the environment, matching the
// precedence in provider.Configure.
func e2eToken() string {
	if v := os.Getenv("TELTONIKA_RMS_TOKEN"); v != "" {
		return v
	}
	return os.Getenv("RMS_ADMIN_TOKEN")
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

// e2eParentCompanyID returns the company the E2E resources should be created
// under. RMS requires parent_id on company creation, and the value is tenant
// specific, so it comes from the environment rather than the repository.
func e2eParentCompanyID() string {
	if v := os.Getenv("RMS_PARENT_COMPANY_ID"); v != "" {
		return v
	}
	return "0"
}
