package provider

import (
	"context"
	"fmt"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CompanyDataSource{}

func NewCompanyDataSource() datasource.DataSource {
	return &CompanyDataSource{}
}

type CompanyDataSource struct {
	client *api.Client
}

type CompanyDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CompanyName types.String `tfsdk:"company_name"`
	ParentID    types.Int64  `tfsdk:"parent_id"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (d *CompanyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_company"
}

func (d *CompanyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a single Teltonika RMS Company.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID.",
			},
			"company_name": schema.StringAttribute{
				Computed:    true,
				Description: "The company name.",
			},
			"parent_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The parent company ID.",
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of devices.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp.",
			},
		},
	}
}

func (d *CompanyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *CompanyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CompanyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := d.client.Get(ctx, fmt.Sprintf("/companies/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading company", fmt.Sprintf("Could not read company %d: %s", data.ID.ValueInt64(), err))
		return
	}

	if companyName, ok := result["company_name"].(string); !ok {
		resp.Diagnostics.AddError(
			"Error parsing company name",
			"Could not parse company_name from API response",
		)
		return
	} else {
		data.CompanyName = types.StringValue(companyName)
	}

	if parentID, ok := result["parent_id"].(float64); ok {
		data.ParentID = types.Int64Value(int64(parentID))
	}

	if deviceCount, ok := result["device_count"].(float64); ok {
		data.DeviceCount = types.Int64Value(int64(deviceCount))
	}

	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
