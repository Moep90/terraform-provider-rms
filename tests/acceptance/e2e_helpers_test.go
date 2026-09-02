package acceptance

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// e2eListPageSize is the page size used when scanning an RMS collection. RMS
// defaults to 10 results per request, which would hide most objects.
const e2eListPageSize = 100

// e2eRunID scopes the objects one E2E run creates. A run that aborts mid-apply
// leaves debris carrying this id, and later runs do not collide with it.
var e2eRunID = strconv.FormatInt(time.Now().Unix(), 10)

// e2eRunPrefix is the name prefix every object an E2E run creates carries.
var e2eRunPrefix = "tfe2e-" + e2eRunID

// e2eClient builds an API client from the same environment provider.Configure
// reads, so the harness and the provider under test address the same tenant.
func e2eClient() *api.Client {
	token := e2eToken()

	// With no override, defer to whatever default the client itself defines,
	// so the harness and the provider always agree on the host.
	baseURL := os.Getenv("TELTONIKA_RMS_BASE_URL")
	if baseURL == "" {
		return api.NewClient(context.Background(), token)
	}

	return api.NewClientWithOptions(context.Background(), token, baseURL, api.Timeout, api.MaxRetries, "e2e")
}

// e2eLookup reads one object out of RMS from the attributes Terraform recorded
// for it. It returns nil while the object exists and api.ErrNotFound once it is
// gone.
type e2eLookup func(ctx context.Context, client *api.Client, attrs map[string]string) error

// e2eReadByID reads the object at path/{id}.
func e2eReadByID(path string) e2eLookup {
	return func(ctx context.Context, client *api.Client, attrs map[string]string) error {
		var result map[string]interface{}
		return client.Get(ctx, path+"/"+attrs["id"], nil, &result)
	}
}

// e2eFindInList scans a collection for the recorded id. RMS defines no GET on
// /vpn/hubs/{id}, /vpn/hubs/routes/{id}, /email-configurations/{id} or
// /users/invitations/{id}, so those objects are confirmed through their
// collection instead. Attribute names passed in forward are sent on as query
// parameters, for the collections that require a filter.
func e2eFindInList(path string, forward ...string) e2eLookup {
	return func(ctx context.Context, client *api.Client, attrs map[string]string) error {
		for offset := 0; ; offset += e2eListPageSize {
			params := map[string]string{
				"limit":  strconv.Itoa(e2eListPageSize),
				"offset": strconv.Itoa(offset),
			}
			for _, name := range forward {
				params[name] = attrs[name]
			}

			var items []map[string]interface{}
			if err := client.Get(ctx, path, params, &items); err != nil {
				return err
			}

			for _, item := range items {
				if id, ok := item["id"].(float64); ok && strconv.FormatInt(int64(id), 10) == attrs["id"] {
					return nil
				}
			}

			if len(items) < e2eListPageSize {
				return api.ErrNotFound
			}
		}
	}
}

// e2eCheckExists fails unless RMS returns the object Terraform recorded for
// resourceName. State on its own proves nothing: the provider can record an id
// for an object the API never created.
func e2eCheckExists(resourceName string, lookup e2eLookup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s: not in Terraform state", resourceName)
		}

		if err := lookup(context.Background(), e2eClient(), rs.Primary.Attributes); err != nil {
			return fmt.Errorf("%s (id %s): RMS does not return the object apply recorded: %w", resourceName, rs.Primary.ID, err)
		}

		return nil
	}
}

// e2eCheckDestroyed fails if RMS still returns any object of resourceType after
// destroy.
func e2eCheckDestroyed(resourceType string, lookup e2eLookup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := e2eClient()

		for name, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			err := lookup(context.Background(), client, rs.Primary.Attributes)
			switch {
			case errors.Is(err, api.ErrNotFound):
			case err != nil:
				return fmt.Errorf("%s (id %s): could not confirm destruction: %w", name, rs.Primary.ID, err)
			default:
				return fmt.Errorf("%s (id %s): still in RMS after destroy", name, rs.Primary.ID)
			}
		}

		return nil
	}
}

// e2eCompanyConfig is the parent company every E2E resource that needs one is
// created under. RMS scopes most objects to a company, and the run creates its
// own rather than touching anything that already exists in the tenant.
func e2eCompanyConfig(name string) string {
	return fmt.Sprintf(`
resource "rms_company" "test" {
  company_name = %q
  parent_id    = %s
}
`, name, e2eParentCompanyID())
}
