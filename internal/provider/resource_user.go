package provider

import (
	"context"
	"fmt"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *api.Client
}

type UserResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CompanyID types.Int64  `tfsdk:"company_id"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS User.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the user.",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The email address.",
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The user role.",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID.",
			},
		},
	}
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Verified against the live RMS API: POST /users returns 405 BAD_REQUEST_METHOD.
	// Creating this resource cannot succeed, so fail here rather than issuing a
	// request whose failure would surface as a confusing parse error, or worse,
	// leave an object behind that Terraform never records.
	resp.Diagnostics.AddError(
		"rms_user cannot be created",
		"The RMS v3 API does not allow POST on /users. Users are created by invitation, so use the rms_invitation resource instead. "+
			"The provider fails here deliberately. Use `terraform import` to manage "+
			"an object that already exists.",
	)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/users/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading user", fmt.Sprintf("Could not read user %d: %s", data.ID.ValueInt64(), err))
		return
	}

	username, ok := result["username"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing username",
			"Could not parse username from API response",
		)
		return
	}
	data.Username = types.StringValue(username)
	email, ok := result["email"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing email",
			"Could not parse email from API response",
		)
		return
	}
	data.Email = types.StringValue(email)
	role, ok := result["role"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing role",
			"Could not parse role from API response",
		)
		return
	}
	data.Role = types.StringValue(role)
	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{
		"email": data.Email.ValueString(),
		"role":  data.Role.ValueString(),
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/users/%d", data.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating user", fmt.Sprintf("Could not update user %d: %s", data.ID.ValueInt64(), err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/users/%d", data.ID.ValueInt64()), nil); err != nil {
		resp.Diagnostics.AddError("Error deleting user", fmt.Sprintf("Could not delete user %d: %s", data.ID.ValueInt64(), err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Configure sets the client for the resource.
func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}
