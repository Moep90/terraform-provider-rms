package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

type RoleResource struct {
	client *api.Client
}

type RoleResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Description   types.String `tfsdk:"description"`
	CompanyID     types.Int64  `tfsdk:"company_id"`
	PermissionIDs types.Set    `tfsdk:"permission_ids"`
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Role for access control.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the role.",
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Title of the role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the role.",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID this role belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"permission_ids": schema.SetAttribute{
				ElementType: types.Int64Type,
				Required:    true,
				Description: "List of permission IDs assigned to this role.",
			},
		},
	}
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permIDs := make([]int64, 0, len(plan.PermissionIDs.Elements()))
	resp.Diagnostics.Append(plan.PermissionIDs.ElementsAs(ctx, &permIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to []interface{} for API
	apiPermIDs := make([]interface{}, len(permIDs))
	for i, id := range permIDs {
		apiPermIDs[i] = int(id)
	}

	createReq := map[string]interface{}{
		"title":        plan.Title.ValueString(),
		"company_id":   []int{int(plan.CompanyID.ValueInt64())},
		"permission_id": apiPermIDs,
	}

	if !plan.Description.IsNull() {
		createReq["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/roles", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating role", fmt.Sprintf("Could not create role: %s", err))
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

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/roles/%d", state.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading role", fmt.Sprintf("Could not read role %d: %s", state.ID.ValueInt64(), err))
		return
	}

	if title, ok := result["title"].(string); ok {
		state.Title = types.StringValue(title)
	}
	if description, ok := result["description"].(string); ok {
		state.Description = types.StringValue(description)
	}
	if companyID, ok := result["company_id"].(float64); ok {
		state.CompanyID = types.Int64Value(int64(companyID))
	}

	// Parse permission_ids from response
	if permIDsRaw, ok := result["permission_id"].([]interface{}); ok {
		permIDs := make([]int64, 0, len(permIDsRaw))
		for _, pid := range permIDsRaw {
			if f, ok := pid.(float64); ok {
				permIDs = append(permIDs, int64(f))
			}
		}
		permSet, diag := types.SetValueFrom(ctx, types.Int64Type, permIDs)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.PermissionIDs = permSet
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permIDs := make([]int64, 0, len(plan.PermissionIDs.Elements()))
	resp.Diagnostics.Append(plan.PermissionIDs.ElementsAs(ctx, &permIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPermIDs := make([]interface{}, len(permIDs))
	for i, id := range permIDs {
		apiPermIDs[i] = int(id)
	}

	updateReq := map[string]interface{}{
		"permission_id": apiPermIDs,
	}

	if !plan.Description.IsNull() {
		updateReq["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/roles/%d", plan.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating role", fmt.Sprintf("Could not update role %d: %s", plan.ID.ValueInt64(), err))
		return
	}

	if title, ok := result["title"].(string); ok {
		plan.Title = types.StringValue(title)
	}
	if description, ok := result["description"].(string); ok {
		plan.Description = types.StringValue(description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/roles/%d", state.ID.ValueInt64()), nil); err != nil {
		resp.Diagnostics.AddError("Error deleting role", fmt.Sprintf("Could not delete role %d: %s", state.ID.ValueInt64(), err))
		return
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
