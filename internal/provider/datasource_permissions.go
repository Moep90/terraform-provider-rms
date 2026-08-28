package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &PermissionsDataSource{}

func NewPermissionsDataSource() datasource.DataSource {
	return &PermissionsDataSource{}
}

type PermissionsDataSource struct {
	client *api.Client
}

type PermissionsDataSourceModel struct {
	ID          types.String     `tfsdk:"id"`
	Permissions []PermissionItem `tfsdk:"permissions"`
}

type PermissionItem struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *PermissionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_permissions"
}

func (d *PermissionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all available Teltonika RMS permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"permissions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of available permissions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Permission name/identifier.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Permission description.",
						},
					},
				},
			},
		},
	}
}

func (d *PermissionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PermissionsDataSourceModel

	var result []map[string]interface{}

	if err := d.client.Get(ctx, "/permissions", nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading permissions", fmt.Sprintf("Could not read permissions: %s", err))
		return
	}

	var permissions []PermissionItem
	for _, p := range result {
		name, ok := p["name"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing permission name", "Could not parse name from API response")
			return
		}

		description := ""
		if desc, ok := p["description"].(string); ok {
			description = desc
		}

		permissions = append(permissions, PermissionItem{
			Name:        types.StringValue(name),
			Description: types.StringValue(description),
		})
	}

	data.ID = types.StringValue("permissions-data-source")
	data.Permissions = permissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
