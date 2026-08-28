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

var _ resource.Resource = &VPNHubResource{}
var _ resource.ResourceWithImportState = &VPNHubResource{}

func NewVPNHubResource() resource.Resource {
	return &VPNHubResource{}
}

type VPNHubResource struct {
	client *api.Client
}

type VPNHubResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	HubZone     types.String `tfsdk:"hub_zone"`
	VPNType     types.String `tfsdk:"vpn_type"`
	TagIDs      types.Set    `tfsdk:"tag_ids"`
}

func (r *VPNHubResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_vpn_hub"
}

func (r *VPNHubResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS VPN Hub for secure multi-site connectivity.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the VPN hub.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "VPN hub name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "VPN hub description.",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID this VPN hub belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"hub_zone": schema.StringAttribute{
				Required:    true,
				Description: "VPN hub server location (e.g., frankfurt-1, bahrain-1).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpn_type": schema.StringAttribute{
				Optional:    true,
				Description: "VPN type: tap or tun.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tag_ids": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Description: "List of tag IDs assigned to this VPN hub.",
			},
		},
	}
}

func (r *VPNHubResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VPNHubResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"company_id": int(plan.CompanyID.ValueInt64()),
		"hub_zone":   plan.HubZone.ValueString(),
	}

	if !plan.Description.IsNull() {
		createReq["description"] = plan.Description.ValueString()
	}
	if !plan.VPNType.IsNull() {
		createReq["vpn_type"] = plan.VPNType.ValueString()
	}

	if !plan.TagIDs.IsNull() {
		tagIDs := make([]int64, 0, len(plan.TagIDs.Elements()))
		resp.Diagnostics.Append(plan.TagIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiTagIDs := make([]interface{}, len(tagIDs))
		for i, id := range tagIDs {
			apiTagIDs[i] = int(id)
		}
		createReq["tag_id"] = apiTagIDs
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/vpn/hubs", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating VPN hub", fmt.Sprintf("Could not create VPN hub: %s", err))
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

func (r *VPNHubResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VPNHubResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/vpn/hubs/%d", state.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading VPN hub", fmt.Sprintf("Could not read VPN hub %d: %s", state.ID.ValueInt64(), err))
		return
	}

	if name, ok := result["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		state.Description = types.StringValue(description)
	}
	if companyID, ok := result["company_id"].(float64); ok {
		state.CompanyID = types.Int64Value(int64(companyID))
	}
	if hubZone, ok := result["hub_zone"].(string); ok {
		state.HubZone = types.StringValue(hubZone)
	}
	if vpnType, ok := result["vpn_type"].(string); ok {
		state.VPNType = types.StringValue(vpnType)
	}

	// Parse tag_ids from response
	if tagIDsRaw, ok := result["tag_id"].([]interface{}); ok {
		tagIDs := make([]int64, 0, len(tagIDsRaw))
		for _, tid := range tagIDsRaw {
			if f, ok := tid.(float64); ok {
				tagIDs = append(tagIDs, int64(f))
			}
		}
		tagSet, diag := types.SetValueFrom(ctx, types.Int64Type, tagIDs)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.TagIDs = tagSet
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VPNHubResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VPNHubResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !plan.Description.IsNull() {
		updateReq["description"] = plan.Description.ValueString()
	}

	if len(updateReq) == 0 {
		resp.Diagnostics.AddError("No fields to update", "At least one field must be provided for update")
		return
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/vpn/hubs/%d", plan.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating VPN hub", fmt.Sprintf("Could not update VPN hub %d: %s", plan.ID.ValueInt64(), err))
		return
	}

	if name, ok := result["name"].(string); ok {
		plan.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		plan.Description = types.StringValue(description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VPNHubResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VPNHubResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteReq := map[string]interface{}{
		"id": []int{int(state.ID.ValueInt64())},
	}

	r.client.Delete(ctx, "/vpn/hubs", deleteReq) //nolint:errcheck // delete is best-effort
}

func (r *VPNHubResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VPNHubResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
