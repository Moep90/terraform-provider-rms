package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ resource.Resource = &DeviceResource{}
var _ resource.ResourceWithImportState = &DeviceResource{}

func NewDeviceResource() resource.Resource {
	return &DeviceResource{}
}

type DeviceResource struct {
	client *api.Client
}

type DeviceResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DeviceSeries     types.String `tfsdk:"device_series"`
	Serial           types.String `tfsdk:"serial"`
	Mac              types.String `tfsdk:"mac"`
	Imei             types.String `tfsdk:"imei"`
	CompanyID        types.Int64  `tfsdk:"company_id"`
	AutoCreditEnable types.Bool   `tfsdk:"auto_credit_enable"`
	Password         types.String `tfsdk:"password"`
	Status           types.String `tfsdk:"status"`
	Firmware         types.String `tfsdk:"firmware"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (r *DeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *DeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Device. Supports RUT, TRB, and other device series.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the device.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the device.",
			},
			"device_series": schema.StringAttribute{
				Required:    true,
				Description: "The device series: rut, trb, etc.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"serial": schema.StringAttribute{
				Required:    true,
				Description: "The device serial number.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac": schema.StringAttribute{
				Optional:    true,
				Description: "The MAC address (required for RUT devices).",
			},
			"imei": schema.StringAttribute{
				Optional:    true,
				Description: "The IMEI (required for TRB devices).",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID to assign the device to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"auto_credit_enable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to automatically enable credits for the device.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The device password for initial validation.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The device status: online, offline, not_activated.",
			},
			"firmware": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The device firmware version.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the device was added.",
			},
		},
	}
}

func (r *DeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeviceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":               data.Name.ValueString(),
		"device_series":      data.DeviceSeries.ValueString(),
		"serial":             data.Serial.ValueString(),
		"company_id":         data.CompanyID.ValueInt64(),
		"auto_credit_enable": data.AutoCreditEnable.ValueBool(),
	}

	if !data.Mac.IsNull() {
		createReq["mac"] = data.Mac.ValueString()
	}
	if !data.Imei.IsNull() {
		createReq["imei"] = data.Imei.ValueString()
	}
	if !data.Password.IsNull() {
		createReq["password_confirmation"] = data.Password.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/devices", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating device", fmt.Sprintf("Could not create device: %s", err))
		return
	}

	id, ok := result["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse ID from API response",
		)
		return
	}
	data.ID = types.Int64Value(int64(id))
	status, ok := result["status"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing status",
			"Could not parse status from API response",
		)
		return
	}
	data.Status = types.StringValue(status)

	if firmware, ok := result["firmware"].(string); ok {
		data.Firmware = types.StringValue(firmware)
	}

	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeviceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/devices/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading device", fmt.Sprintf("Could not read device %d: %s", data.ID.ValueInt64(), err))
		return
	}

	name, ok := result["name"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing name",
			"Could not parse name from API response",
		)
		return
	}
	data.Name = types.StringValue(name)

	deviceSeries, ok := result["device_series"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing device_series",
			"Could not parse device_series from API response",
		)
		return
	}
	data.DeviceSeries = types.StringValue(deviceSeries)

	if serial, ok := result["serial"].(string); ok {
		data.Serial = types.StringValue(serial)
	}

	if mac, ok := result["mac"].(string); ok {
		data.Mac = types.StringValue(mac)
	}

	if imei, ok := result["imei"].(string); ok {
		data.Imei = types.StringValue(imei)
	}

	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}

	if autoCreditEnable, ok := result["auto_credit_enable"].(bool); ok {
		data.AutoCreditEnable = types.BoolValue(autoCreditEnable)
	}

	status, ok := result["status"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing status",
			"Could not parse status from API response",
		)
		return
	}
	data.Status = types.StringValue(status)

	if firmware, ok := result["firmware"].(string); ok {
		data.Firmware = types.StringValue(firmware)
	}

	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeviceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{
		"name": data.Name.ValueString(),
	}

	if !data.Mac.IsNull() {
		updateReq["mac"] = data.Mac.ValueString()
	}
	if !data.Imei.IsNull() {
		updateReq["imei"] = data.Imei.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/devices/%d", data.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating device", fmt.Sprintf("Could not update device %d: %s", data.ID.ValueInt64(), err))
		return
	}

	var updatedResult map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/devices/%d", data.ID.ValueInt64()), nil, &updatedResult); err != nil {
		resp.Diagnostics.AddError("Error reading device", fmt.Sprintf("Could not read device %d: %s", data.ID.ValueInt64(), err))
		return
	}

	if status, ok := updatedResult["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if firmware, ok := updatedResult["firmware"].(string); ok {
		data.Firmware = types.StringValue(firmware)
	}
	if createdAt, ok := updatedResult["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeviceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/devices/%d", data.ID.ValueInt64()), nil); err != nil {
		resp.Diagnostics.AddError("Error deleting device", fmt.Sprintf("Could not delete device %d: %s", data.ID.ValueInt64(), err))
		return
	}
}

func (r *DeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Configure sets the client for the resource.
func (r *DeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
