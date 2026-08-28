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

var _ resource.Resource = &EmailConfigurationResource{}
var _ resource.ResourceWithImportState = &EmailConfigurationResource{}

func NewEmailConfigurationResource() resource.Resource {
	return &EmailConfigurationResource{}
}

type EmailConfigurationResource struct {
	client *api.Client
}

type EmailConfigurationResourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Email    types.String `tfsdk:"email"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r *EmailConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_email_configuration"
}

func (r *EmailConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Email Configuration for alert notifications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the email configuration.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Email name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				Required:    true,
				Description: "SMTP host name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "SMTP port number.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Email address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "SMTP username.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SMTP password.",
			},
		},
	}
}

func (r *EmailConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EmailConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":     plan.Name.ValueString(),
		"host":     plan.Host.ValueString(),
		"port":     plan.Port.ValueInt64(),
		"email":    plan.Email.ValueString(),
		"username": plan.Username.ValueString(),
	}

	if !plan.Password.IsNull() {
		createReq["password"] = plan.Password.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/email-configurations", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating email configuration", fmt.Sprintf("Could not create email configuration: %s", err))
		return
	}

	id, ok := result["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Error parsing ID", "Could not parse ID from API response")
		return
	}
	plan.ID = types.Int64Value(int64(id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EmailConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EmailConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/email-configurations/%d", state.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading email configuration", fmt.Sprintf("Could not read email configuration %d: %s", state.ID.ValueInt64(), err))
		return
	}

	if name, ok := result["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if host, ok := result["host"].(string); ok {
		state.Host = types.StringValue(host)
	}
	if port, ok := result["port"].(float64); ok {
		state.Port = types.Int64Value(int64(port))
	}
	if email, ok := result["email"].(string); ok {
		state.Email = types.StringValue(email)
	}
	if username, ok := result["username"].(string); ok {
		state.Username = types.StringValue(username)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EmailConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EmailConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{
		"name":     plan.Name.ValueString(),
		"host":     plan.Host.ValueString(),
		"port":     plan.Port.ValueInt64(),
		"email":    plan.Email.ValueString(),
		"username": plan.Username.ValueString(),
	}

	if !plan.Password.IsNull() {
		updateReq["password"] = plan.Password.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/email-configurations/%d", plan.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating email configuration", fmt.Sprintf("Could not update email configuration %d: %s", plan.ID.ValueInt64(), err))
		return
	}

	if name, ok := result["name"].(string); ok {
		plan.Name = types.StringValue(name)
	}
	if host, ok := result["host"].(string); ok {
		plan.Host = types.StringValue(host)
	}
	if port, ok := result["port"].(float64); ok {
		plan.Port = types.Int64Value(int64(port))
	}
	if email, ok := result["email"].(string); ok {
		plan.Email = types.StringValue(email)
	}
	if username, ok := result["username"].(string); ok {
		plan.Username = types.StringValue(username)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EmailConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EmailConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/email-configurations/%d", state.ID.ValueInt64()), nil); err != nil {
		resp.Diagnostics.AddError("Error deleting email configuration", fmt.Sprintf("Could not delete email configuration %d: %s", state.ID.ValueInt64(), err))
		return
	}
}

func (r *EmailConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *EmailConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
