package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ resource.Resource = &VPNHubRouteResource{}
var _ resource.ResourceWithImportState = &VPNHubRouteResource{}

func NewVPNHubRouteResource() resource.Resource {
	return &VPNHubRouteResource{}
}

type VPNHubRouteResource struct {
	client *api.Client
}

type VPNHubRouteResourceModel struct {
	ID            types.String `tfsdk:"id"`
	VPNHubID      types.Int64  `tfsdk:"vpn_hub_id"`
	IPAddress     types.String `tfsdk:"ip_address"`
	Netmask       types.String `tfsdk:"netmask"`
	VPNHubUserID  types.Int64  `tfsdk:"vpn_hub_user_id"`
	Description   types.String `tfsdk:"description"`
}

func (r *VPNHubRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_vpn_hub_route"
}

func (r *VPNHubRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS VPN Hub Route for routing configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the route (format: vpn_hub_id:vpn_hub_user_id:ip_address).",
			},
			"vpn_hub_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the VPN hub.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"vpn_hub_user_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the VPN hub user.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ip_address": schema.StringAttribute{
				Required:    true,
				Description: "The IP address for the route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"netmask": schema.StringAttribute{
				Required:    true,
				Description: "The netmask for the route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description for the route.",
			},
		},
	}
}

func (r *VPNHubRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VPNHubRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"vpn_hub_id":       plan.VPNHubID.ValueInt64(),
		"vpn_hub_user_id":  plan.VPNHubUserID.ValueInt64(),
		"ip_address":       plan.IPAddress.ValueString(),
		"netmask":          plan.Netmask.ValueString(),
	}

	if !plan.Description.IsNull() {
		createReq["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/vpn/hubs/routes", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating VPN hub route", fmt.Sprintf("Could not create VPN hub route: %s", err))
		return
	}

	id, ok := result["id"].(string)
	if !ok {
		resp.Diagnostics.AddError("Error parsing ID", "Could not parse ID from API response")
		return
	}
	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VPNHubRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VPNHubRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/vpn/hubs/routes/%s", state.ID.ValueString()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading VPN hub route", fmt.Sprintf("Could not read VPN hub route %s: %s", state.ID.ValueString(), err))
		return
	}

	if vpnHubID, ok := result["vpn_hub_id"].(float64); ok {
		state.VPNHubID = types.Int64Value(int64(vpnHubID))
	}
	if vpnHubUserID, ok := result["vpn_hub_user_id"].(float64); ok {
		state.VPNHubUserID = types.Int64Value(int64(vpnHubUserID))
	}
	if ipAddress, ok := result["ip_address"].(string); ok {
		state.IPAddress = types.StringValue(ipAddress)
	}
	if netmask, ok := result["netmask"].(string); ok {
		state.Netmask = types.StringValue(netmask)
	}
	if description, ok := result["description"].(string); ok {
		state.Description = types.StringValue(description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VPNHubRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VPNHubRouteResourceModel

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
	if err := r.client.Put(ctx, fmt.Sprintf("/vpn/hubs/routes/%s", plan.ID.ValueString()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating VPN hub route", fmt.Sprintf("Could not update VPN hub route %s: %s", plan.ID.ValueString(), err))
		return
	}

	if description, ok := result["description"].(string); ok {
		plan.Description = types.StringValue(description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VPNHubRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VPNHubRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteReq := map[string]interface{}{
		"vpn_hub_id":      int(state.VPNHubID.ValueInt64()),
		"vpn_hub_user_id": int(state.VPNHubUserID.ValueInt64()),
		"ip_address":      state.IPAddress.ValueString(),
		"netmask":         state.Netmask.ValueString(),
	}

	r.client.Delete(ctx, "/vpn/hubs/routes", deleteReq) //nolint:errcheck // delete is best-effort
}

func (r *VPNHubRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: vpn_hub_id:vpn_hub_user_id:ip_address",
		)
		return
	}

	vpnHubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid VPN Hub ID", fmt.Sprintf("Cannot parse vpn_hub_id: %s", err))
		return
	}

	vpnHubUserID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid VPN Hub User ID", fmt.Sprintf("Cannot parse vpn_hub_user_id: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &VPNHubRouteResourceModel{
		ID:           types.StringValue(req.ID),
		VPNHubID:     types.Int64Value(vpnHubID),
		VPNHubUserID: types.Int64Value(vpnHubUserID),
		IPAddress:    types.StringValue(parts[2]),
	})...)
}

func (r *VPNHubRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
