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
	RoleID      types.Int64      `tfsdk:"role_id"`
	Permissions []PermissionItem `tfsdk:"permissions"`
}

type PermissionItem struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Category    types.String `tfsdk:"category"`
}

func (d *PermissionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_permissions"
}

func (d *PermissionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the permissions assigned to a Teltonika RMS role. RMS exposes " +
			"permissions per role rather than as a global catalogue, so role_id is required. " +
			"The built-in Administrator role holds every permission and is the usual choice " +
			"when discovering permission IDs for rms_role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"role_id": schema.Int64Attribute{
				Required:    true,
				Description: "ID of the role whose permissions should be read.",
			},
			"permissions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Permissions assigned to the role, sorted by name.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Permission ID, as accepted by rms_role.permission_ids.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Permission slug, for example view_pending_device_actions.",
						},
						"title": schema.StringAttribute{
							Computed:    true,
							Description: "Human readable permission title.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Permission description.",
						},
						"category": schema.StringAttribute{
							Computed:    true,
							Description: "Category the permission belongs to, for example Device actions.",
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

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := data.RoleID.ValueInt64()

	permissions, err := readRolePermissions(ctx, d.client, roleID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading permissions", fmt.Sprintf("Could not read permissions for role %d: %s", roleID, err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("role-%d-permissions", roleID))
	data.Permissions = permissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readRolePermissions reads /roles/{id}/permissions. RMS has no global
// /permissions endpoint; permissions are only reachable through a role.
func readRolePermissions(ctx context.Context, client *api.Client, roleID int64) ([]PermissionItem, error) {
	var result []map[string]interface{}

	if err := client.Get(ctx, fmt.Sprintf("/roles/%d/permissions", roleID), nil, &result); err != nil {
		return nil, err
	}

	permissions := make([]PermissionItem, 0, len(result))
	for i, p := range result {
		id, ok := p["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("permission at index %d has no numeric id (got %T: %v)", i, p["id"], p["id"])
		}

		name, ok := p["name"].(string)
		if !ok {
			return nil, fmt.Errorf("permission %d has no string name (got %T: %v)", int64(id), p["name"], p["name"])
		}

		item := PermissionItem{
			ID:          types.Int64Value(int64(id)),
			Name:        types.StringValue(name),
			Title:       types.StringValue(""),
			Description: types.StringValue(""),
			Category:    types.StringValue(""),
		}
		if v, ok := p["title"].(string); ok {
			item.Title = types.StringValue(v)
		}
		if v, ok := p["description"].(string); ok {
			item.Description = types.StringValue(v)
		}
		if v, ok := p["category"].(string); ok {
			item.Category = types.StringValue(v)
		}

		permissions = append(permissions, item)
	}

	// The API does not guarantee a stable order, but list attributes are
	// addressed by index in state.
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Name.ValueString() < permissions[j].Name.ValueString()
	})

	return permissions, nil
}

// rolePermissionIDs returns just the permission IDs assigned to a role.
func rolePermissionIDs(ctx context.Context, client *api.Client, roleID int64) ([]int64, error) {
	permissions, err := readRolePermissions(ctx, client, roleID)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(permissions))
	for _, p := range permissions {
		ids = append(ids, p.ID.ValueInt64())
	}

	return ids, nil
}
