package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	ID           types.String `tfsdk:"id"`
	VPNHubID     types.Int64  `tfsdk:"vpn_hub_id"`
	IPAddress    types.String `tfsdk:"ip_address"`
	Netmask      types.String `tfsdk:"netmask"`
	VPNHubUserID types.Int64  `tfsdk:"vpn_hub_user_id"`
	Description  types.String `tfsdk:"description"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
		"vpn_hub_id":      plan.VPNHubID.ValueInt64(),
		"vpn_hub_user_id": plan.VPNHubUserID.ValueInt64(),
		"ip_address":      plan.IPAddress.ValueString(),
		"netmask":         plan.Netmask.ValueString(),
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

	// RMS exposes no GET /vpn/hubs/routes/{id}; routes are read off the
	// collection, scoped to the hub and hub user.
	params := map[string]string{
		"vpn_hub_id":      strconv.FormatInt(state.VPNHubID.ValueInt64(), 10),
		"vpn_hub_user_id": strconv.FormatInt(state.VPNHubUserID.ValueInt64(), 10),
	}

	var payload json.RawMessage
	if err := r.client.Get(ctx, "/vpn/hubs/routes", params, &payload); err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading VPN hub route", fmt.Sprintf("Could not read VPN hub route %s: %s", state.ID.ValueString(), err))
		return
	}

	var routes []map[string]interface{}
	if err := json.Unmarshal(payload, &routes); err != nil {
		// RMS answered with a Status API channel handle rather than a route
		// list. Nothing can be reconciled, and treating that as "route gone"
		// would destroy and recreate the route on every refresh, so state is
		// left untouched.
		tflog.Warn(ctx, "VPN hub route list unavailable, keeping state", map[string]interface{}{
			"id": state.ID.ValueString(),
		})
		return
	}

	// Every attribute of this resource is a RequiresReplace input, so the only
	// thing the collection can tell us is whether the route is still there.
	for _, route := range routes {
		ip, ok := route["ip"].(string)
		if !ok {
			ip, ok = route["ip_address"].(string)
		}
		if !ok {
			continue
		}

		netmask, ok := route["netmask"].(string)
		if !ok {
			continue
		}

		if ip == state.IPAddress.ValueString() && netmask == state.Netmask.ValueString() {
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *VPNHubRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Every attribute requires replacement, so Terraform never plans an
	// in-place update here. RMS exposes no update operation for routes.
	resp.Diagnostics.AddError(
		"rms_vpn_hub_route cannot be updated",
		"The RMS API exposes no update operation for VPN hub routes. Every "+
			"attribute of this resource requires replacement instead.",
	)
}

func (r *VPNHubRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VPNHubRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The path segment carries the hub id; the route itself is selected through
	// the request body.
	deleteReq := map[string]interface{}{
		"vpn_hub_user_id": state.VPNHubUserID.ValueInt64(),
		"ip_address":      state.IPAddress.ValueString(),
		"netmask":         state.Netmask.ValueString(),
	}

	deletePath := fmt.Sprintf("/vpn/hubs/routes/%d", state.VPNHubID.ValueInt64())
	if err := r.client.DeleteWithBody(ctx, deletePath, deleteReq, nil); err != nil {
		if errors.Is(err, api.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting VPN hub route", fmt.Sprintf("Could not delete VPN hub route: %s", err))
		return
	}
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
