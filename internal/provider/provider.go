package provider

import (
	"context"
	"os"
	"time"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TeltonikaProvider satisfies various provider interfaces.
var _ provider.Provider = &TeltonikaProvider{}
var _ provider.ProviderWithFunctions = &TeltonikaProvider{}

// TeltonikaProvider is the provider implementation.
type TeltonikaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// tests.
	version string
}

// TeltonikaProviderModel describes the provider data model.
type TeltonikaProviderModel struct {
	Token    types.String `tfsdk:"token"`
	BaseURL  types.String `tfsdk:"base_url"`
	Timeout  types.Int64  `tfsdk:"timeout"`
	MaxRetry types.Int64  `tfsdk:"max_retry"`
}

// Metadata sets the provider's name and version.
func (p *TeltonikaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rms"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *TeltonikaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Teltonika RMS companies, devices, users, roles, tasks and VPN hubs. " +
			"Configure the provider with an API token, either inline or via the " +
			"TELTONIKA_RMS_TOKEN environment variable.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "API token for authentication. Can also be set via TELTONIKA_RMS_TOKEN environment variable.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL for the Teltonika RMS API. Defaults to https://rms.teltonika-networks.com/api",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Request timeout in seconds. Defaults to 30.",
			},
			"max_retry": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of retries for failed requests. Defaults to 3.",
			},
		},
	}
}

// Configure prepares a Teltonika API client for data sources and resources.
func (p *TeltonikaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TeltonikaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	if data.Token.IsNull() {
		token := os.Getenv("TELTONIKA_RMS_TOKEN")
		if token != "" {
			data.Token = types.StringValue(token)
		}
	}

	if data.BaseURL.IsNull() {
		baseURL := os.Getenv("TELTONIKA_RMS_BASE_URL")
		if baseURL == "" {
			baseURL = api.BaseURL
		}
		data.BaseURL = types.StringValue(baseURL)
	}

	if data.Timeout.IsNull() {
		data.Timeout = types.Int64Value(30)
	}

	if data.MaxRetry.IsNull() {
		data.MaxRetry = types.Int64Value(3)
	}

	if data.Token.IsNull() || data.Token.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider requires an API token. Set the token attribute or TELTONIKA_RMS_TOKEN environment variable.",
		)
		return
	}

	// Create API client
	client := api.NewClientWithOptions(
		ctx,
		data.Token.ValueString(),
		data.BaseURL.ValueString(),
		time.Duration(data.Timeout.ValueInt64())*time.Second,
		int(data.MaxRetry.ValueInt64()),
		p.version,
	)

	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *TeltonikaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCompaniesDataSource,
		NewCompanyDataSource,
		NewDevicesDataSource,
		NewDeviceDataSource,
		NewDeviceEsimBootstrapDataSource,
		NewDevicesExportDataSource,
		NewTagsDataSource,
		NewUsersDataSource,
		NewInvitationsDataSource,
		NewPermissionsDataSource,
		NewRolesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *TeltonikaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAlertConfigurationResource,
		NewCompanyResource,
		NewDeviceResource,
		NewDeviceTagsResource,
		NewEmailConfigurationResource,
		NewRoleResource,
		NewTagResource,
		NewUserResource,
		NewInvitationResource,
		NewTaskResource,
		NewTaskGroupResource,
		NewVPNHubResource,
		NewVPNHubRouteResource,
	}
}

// Functions returns the list of functions implemented in the provider.
func (p *TeltonikaProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

// New creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TeltonikaProvider{
			version: version,
		}
	}
}
