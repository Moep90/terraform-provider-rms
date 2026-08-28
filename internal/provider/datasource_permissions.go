package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ID          types.Int64  `tfsdk:"id"`
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
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Permission ID, as accepted by rms_role.permission_ids. Null if the API does not report one.",
						},
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

	// The permission catalogue is larger than the limit of 100 the other list
	// data sources use, so request a limit that covers it in one page.
	if err := d.client.Get(ctx, "/permissions", map[string]string{"limit": "1000"}, &result); err != nil {
		resp.Diagnostics.AddError("Error reading permissions", fmt.Sprintf("Could not read permissions: %s", err))
		return
	}

	permissions := make([]PermissionItem, 0, len(result))
	for i, p := range result {
		name, ok := p["name"].(string)
		if !ok {
			resp.Diagnostics.AddError(
				"Error parsing permission name",
				fmt.Sprintf("Permission at index %d has no string name (got %T: %v)", i, p["name"], p["name"]),
			)
			return
		}

		id := types.Int64Null()
		if raw, ok := p["id"].(float64); ok {
			id = types.Int64Value(int64(raw))
		}

		description := ""
		if desc, ok := p["description"].(string); ok {
			description = desc
		}

		permissions = append(permissions, PermissionItem{
			ID:          id,
			Name:        types.StringValue(name),
			Description: types.StringValue(description),
		})
	}

	// The API does not guarantee a stable order, but list attributes are
	// addressed by index in state.
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Name.ValueString() < permissions[j].Name.ValueString()
	})

	data.ID = types.StringValue("permissions-data-source")
	data.Permissions = permissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
