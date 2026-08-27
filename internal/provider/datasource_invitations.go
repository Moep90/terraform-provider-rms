package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &InvitationsDataSource{}

func NewInvitationsDataSource() datasource.DataSource {
	return &InvitationsDataSource{}
}

type InvitationsDataSource struct {
	client *api.Client
}

type InvitationsDataSourceModel struct {
	ID          types.String          `tfsdk:"id"`
	Invitations []InvitationDataModel `tfsdk:"invitations"`
}

type InvitationDataModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CompanyID types.Int64  `tfsdk:"company_id"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *InvitationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invitations"
}

func (d *InvitationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of Teltonika RMS User Invitations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"invitations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of invitations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.Int64Attribute{Computed: true},
						"email":      schema.StringAttribute{Computed: true},
						"role":       schema.StringAttribute{Computed: true},
						"company_id": schema.Int64Attribute{Computed: true},
						"created_at": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *InvitationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InvitationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InvitationsDataSourceModel
	var result []map[string]interface{}

	if err := d.client.Get(ctx, "/users/invitations", map[string]string{"limit": "100"}, &result); err != nil {
		resp.Diagnostics.AddError("Error reading invitations", fmt.Sprintf("Could not read invitations: %s", err))
		return
	}

	var invitations []InvitationDataModel
	for _, inv := range result {
		id, ok := inv["id"].(float64)
		if !ok {
			resp.Diagnostics.AddError("Error parsing invitation ID", "Could not parse id from API response")
			return
		}
		email, ok := inv["email"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing invitation email", "Could not parse email from API response")
			return
		}
		role, ok := inv["role"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing invitation role", "Could not parse role from API response")
			return
		}
		invitation := InvitationDataModel{
			ID:    types.Int64Value(int64(id)),
			Email: types.StringValue(email),
			Role:  types.StringValue(role),
		}
		if companyID, ok := inv["company_id"].(float64); ok {
			invitation.CompanyID = types.Int64Value(int64(companyID))
		}
		if createdAt, ok := inv["created_at"].(string); ok {
			invitation.CreatedAt = types.StringValue(createdAt)
		}
		invitations = append(invitations, invitation)
	}

	data.ID = types.StringValue("invitations-data-source")
	data.Invitations = invitations
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
