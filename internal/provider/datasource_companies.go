package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &CompaniesDataSource{}

func NewCompaniesDataSource() datasource.DataSource {
	return &CompaniesDataSource{}
}

type CompaniesDataSource struct {
	client *api.Client
}

type CompaniesDataSourceModel struct {
	ID        types.String       `tfsdk:"id"`
	Companies []CompanyDataModel `tfsdk:"companies"`
}

type CompanyDataModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CompanyName types.String `tfsdk:"company_name"`
	ParentID    types.Int64  `tfsdk:"parent_id"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
}

func (d *CompaniesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_companies"
}

func (d *CompaniesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all Teltonika RMS Companies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"companies": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of companies.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Company ID.",
						},
						"company_name": schema.StringAttribute{
							Computed:    true,
							Description: "Company name.",
						},
						"parent_id": schema.Int64Attribute{
							Computed:    true,
							Description: "Parent company ID.",
						},
						"device_count": schema.Int64Attribute{
							Computed:    true,
							Description: "Number of devices.",
						},
					},
				},
			},
		},
	}
}

func (d *CompaniesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CompaniesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CompaniesDataSourceModel

	var companies []CompanyDataModel
	var result []map[string]interface{}

	if err := d.client.Get(ctx, "/companies", map[string]string{"limit": "100"}, &result); err != nil {
		resp.Diagnostics.AddError("Error reading companies", fmt.Sprintf("Could not read companies: %s", err))
		return
	}

	for _, c := range result {
		id, ok := c["id"].(float64)
		if !ok {
			resp.Diagnostics.AddError("Error parsing company ID", "Could not parse id from API response")
			return
		}
		companyName, ok := c["company_name"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing company name", "Could not parse company_name from API response")
			return
		}
		company := CompanyDataModel{
			ID:          types.Int64Value(int64(id)),
			CompanyName: types.StringValue(companyName),
		}

		if parentID, ok := c["parent_id"].(float64); ok {
			company.ParentID = types.Int64Value(int64(parentID))
		}

		if deviceCount, ok := c["device_count"].(float64); ok {
			company.DeviceCount = types.Int64Value(int64(deviceCount))
		}

		companies = append(companies, company)
	}

	data.ID = types.StringValue("companies-data-source")
	data.Companies = companies

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
