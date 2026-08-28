package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *api.Client
}

type UsersDataSourceModel struct {
	ID    types.String    `tfsdk:"id"`
	Users []UserDataModel `tfsdk:"users"`
}

type UserDataModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CompanyID types.Int64  `tfsdk:"company_id"`
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of Teltonika RMS Users.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.Int64Attribute{Computed: true},
						"username":   schema.StringAttribute{Computed: true},
						"email":      schema.StringAttribute{Computed: true},
						"role":       schema.StringAttribute{Computed: true},
						"company_id": schema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel
	var result []map[string]interface{}

	if err := d.client.Get(ctx, "/users", map[string]string{"limit": "100"}, &result); err != nil {
		resp.Diagnostics.AddError("Error reading users", fmt.Sprintf("Could not read users: %s", err))
		return
	}

	var users []UserDataModel
	for _, u := range result {
		id, ok := u["id"].(float64)
		if !ok {
			resp.Diagnostics.AddError("Error parsing user ID", "Could not parse id from API response")
			return
		}
		username, ok := u["username"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing username", "Could not parse username from API response")
			return
		}
		email, ok := u["email"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing email", "Could not parse email from API response")
			return
		}
		role, ok := u["role"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing role", "Could not parse role from API response")
			return
		}
		user := UserDataModel{
			ID:       types.Int64Value(int64(id)),
			Username: types.StringValue(username),
			Email:    types.StringValue(email),
			Role:     types.StringValue(role),
		}
		if companyID, ok := u["company_id"].(float64); ok {
			user.CompanyID = types.Int64Value(int64(companyID))
		}
		users = append(users, user)
	}

	data.ID = types.StringValue("users-data-source")
	data.Users = users
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
