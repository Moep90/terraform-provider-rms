package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &RolesDataSource{}

func NewRolesDataSource() datasource.DataSource {
	return &RolesDataSource{}
}

type RolesDataSource struct {
	client *api.Client
}

type RolesDataSourceModel struct {
	ID    types.String   `tfsdk:"id"`
	Roles []RoleDataItem `tfsdk:"roles"`
}

type RoleDataItem struct {
	ID            types.Int64  `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Description   types.String `tfsdk:"description"`
	CompanyID     types.Int64  `tfsdk:"company_id"`
	PermissionIDs types.Set    `tfsdk:"permission_ids"`
}

func (d *RolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_roles"
}

func (d *RolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all Teltonika RMS Roles.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"roles": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Role ID.",
						},
						"title": schema.StringAttribute{
							Computed:    true,
							Description: "Role title.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Role description.",
						},
						"company_id": schema.Int64Attribute{
							Computed:    true,
							Description: "Company ID.",
						},
						"permission_ids": schema.SetAttribute{
							ElementType: types.Int64Type,
							Computed:    true,
							Description: "List of permission IDs assigned to this role.",
						},
					},
				},
			},
		},
	}
}

func (d *RolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RolesDataSourceModel

	var result []map[string]interface{}

	if err := d.client.Get(ctx, "/roles", map[string]string{"limit": "1000"}, &result); err != nil {
		resp.Diagnostics.AddError("Error reading roles", fmt.Sprintf("Could not read roles: %s", err))
		return
	}

	roles := make([]RoleDataItem, 0, len(result))
	for i, r := range result {
		id, ok := r["id"].(float64)
		if !ok {
			resp.Diagnostics.AddError(
				"Error parsing role ID",
				fmt.Sprintf("Role at index %d has no numeric id (got %T: %v)", i, r["id"], r["id"]),
			)
			return
		}

		title, ok := r["title"].(string)
		if !ok {
			resp.Diagnostics.AddError(
				"Error parsing role title",
				fmt.Sprintf("Role %d has no string title (got %T: %v)", int64(id), r["title"], r["title"]),
			)
			return
		}

		description := ""
		if desc, ok := r["description"].(string); ok {
			description = desc
		}

		companyID, err := parseRoleCompanyID(r["company_id"])
		if err != nil {
			resp.Diagnostics.AddError("Error parsing role company ID", fmt.Sprintf("Role %d: %s", int64(id), err))
			return
		}

		permissionIDs, err := parseRolePermissionIDs(r["permission_id"])
		if err != nil {
			resp.Diagnostics.AddError("Error parsing role permissions", fmt.Sprintf("Role %d: %s", int64(id), err))
			return
		}

		permSet, diag := types.SetValueFrom(ctx, types.Int64Type, permissionIDs)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		roles = append(roles, RoleDataItem{
			ID:            types.Int64Value(int64(id)),
			Title:         types.StringValue(title),
			Description:   types.StringValue(description),
			CompanyID:     companyID,
			PermissionIDs: permSet,
		})
	}

	// The API does not guarantee a stable order, but list attributes are
	// addressed by index in state.
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].ID.ValueInt64() < roles[j].ID.ValueInt64()
	})

	data.ID = types.StringValue("roles-data-source")
	data.Roles = roles

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
