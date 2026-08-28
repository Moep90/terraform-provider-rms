package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ resource.Resource = &AlertConfigurationResource{}
var _ resource.ResourceWithImportState = &AlertConfigurationResource{}

func NewAlertConfigurationResource() resource.Resource {
	return &AlertConfigurationResource{}
}

type AlertConfigurationResource struct {
	client *api.Client
}

type AlertConfigurationResourceModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	DeviceID           types.Int64  `tfsdk:"device_id"`
	AlertTypeID        types.Int64  `tfsdk:"alert_type_id"`
	AlertSubtypeID     types.Int64  `tfsdk:"alert_subtype_id"`
	Action             types.Int64  `tfsdk:"action"`
	Subject            types.String `tfsdk:"subject"`
	Message            types.String `tfsdk:"message"`
	Email              types.String `tfsdk:"email"`
	SMTPConfigID       types.Int64  `tfsdk:"smtp_config_id"`
	DeliveryRetry      types.Bool   `tfsdk:"delivery_retry"`
	RetryInterval      types.Int64  `tfsdk:"retry_interval"`
	RetryCount         types.Int64  `tfsdk:"retry_count"`
	RedundancyInterval types.Int64  `tfsdk:"redundancy_interval"`
	DataLimit          types.Int64  `tfsdk:"data_limit"`
	SIM                types.Int64  `tfsdk:"sim"`
}

func (r *AlertConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_alert_configuration"
}

func (r *AlertConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Alert Configuration for device monitoring alerts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the alert configuration.",
			},
			"device_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the device to monitor.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"alert_type_id": schema.Int64Attribute{
				Required:    true,
				Description: "The alert type ID.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"alert_subtype_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The alert subtype ID.",
			},
			"action": schema.Int64Attribute{
				Optional:    true,
				Description: "Alert action (1=email, 2=sms, 3=webhook).",
			},
			"subject": schema.StringAttribute{
				Optional:    true,
				Description: "Subject line for email alerts.",
			},
			"message": schema.StringAttribute{
				Optional:    true,
				Description: "Message content for alerts.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "Email address for notifications.",
			},
			"smtp_config_id": schema.Int64Attribute{
				Optional:    true,
				Description: "SMTP configuration ID for sending emails.",
			},
			"delivery_retry": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether delivery retry is enabled.",
			},
			"retry_interval": schema.Int64Attribute{
				Optional:    true,
				Description: "Retry interval in minutes.",
			},
			"retry_count": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of retries.",
			},
			"redundancy_interval": schema.Int64Attribute{
				Optional:    true,
				Description: "Redundancy interval in minutes.",
			},
			"data_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Data limit interval in minutes.",
			},
			"sim": schema.Int64Attribute{
				Optional:    true,
				Description: "SIM card number (1 or 2).",
			},
		},
	}
}

func (r *AlertConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlertConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertData := map[string]interface{}{
		"device_id":     plan.DeviceID.ValueInt64(),
		"alert_type_id": plan.AlertTypeID.ValueInt64(),
	}

	if !plan.AlertSubtypeID.IsNull() {
		alertData["alert_subtype_id"] = plan.AlertSubtypeID.ValueInt64()
	}
	if !plan.Action.IsNull() {
		alertData["action"] = plan.Action.ValueInt64()
	}
	if !plan.Subject.IsNull() {
		alertData["subject"] = plan.Subject.ValueString()
	}
	if !plan.Message.IsNull() {
		alertData["message"] = plan.Message.ValueString()
	}
	if !plan.Email.IsNull() {
		alertData["email"] = plan.Email.ValueString()
	}
	if !plan.SMTPConfigID.IsNull() {
		alertData["smtp_config_id"] = plan.SMTPConfigID.ValueInt64()
	}
	if !plan.DeliveryRetry.IsNull() {
		alertData["delivery_retry"] = plan.DeliveryRetry.ValueBool()
	}
	if !plan.RetryInterval.IsNull() {
		alertData["retry_interval"] = plan.RetryInterval.ValueInt64()
	}
	if !plan.RetryCount.IsNull() {
		alertData["retry_count"] = plan.RetryCount.ValueInt64()
	}
	if !plan.RedundancyInterval.IsNull() {
		alertData["redundancy_interval"] = plan.RedundancyInterval.ValueInt64()
	}
	if !plan.DataLimit.IsNull() {
		alertData["data_limit"] = plan.DataLimit.ValueInt64()
	}
	if !plan.SIM.IsNull() {
		alertData["sim"] = plan.SIM.ValueInt64()
	}

	createReq := map[string]interface{}{
		"data": []map[string]interface{}{alertData},
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/alerts-configurations", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating alert configuration", fmt.Sprintf("Could not create alert configuration: %s", err))
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

func (r *AlertConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/alerts-configurations/%d", state.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading alert configuration", fmt.Sprintf("Could not read alert configuration %d: %s", state.ID.ValueInt64(), err))
		return
	}

	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if alertData, ok := data[0].(map[string]interface{}); ok {
			if deviceID, ok := alertData["device_id"].(float64); ok {
				state.DeviceID = types.Int64Value(int64(deviceID))
			}
			if alertTypeID, ok := alertData["alert_type_id"].(float64); ok {
				state.AlertTypeID = types.Int64Value(int64(alertTypeID))
			}
			if alertSubtypeID, ok := alertData["alert_subtype_id"].(float64); ok {
				state.AlertSubtypeID = types.Int64Value(int64(alertSubtypeID))
			}
			if action, ok := alertData["action"].(string); ok {
				var actionInt int64
				fmt.Sscanf(action, "%d", &actionInt) //nolint:errcheck // parsing string to int
				state.Action = types.Int64Value(actionInt)
			}
			if subject, ok := alertData["subject"].(string); ok {
				state.Subject = types.StringValue(subject)
			}
			if message, ok := alertData["message"].(string); ok {
				state.Message = types.StringValue(message)
			}
			if email, ok := alertData["email"].(string); ok {
				state.Email = types.StringValue(email)
			}
			if smtpConfigID, ok := alertData["smtp_config_id"].(float64); ok {
				state.SMTPConfigID = types.Int64Value(int64(smtpConfigID))
			}
			if deliveryRetry, ok := alertData["delivery_retry"].(bool); ok {
				state.DeliveryRetry = types.BoolValue(deliveryRetry)
			}
			if retryInterval, ok := alertData["retry_interval"].(float64); ok {
				state.RetryInterval = types.Int64Value(int64(retryInterval))
			}
			if retryCount, ok := alertData["retry_count"].(float64); ok {
				state.RetryCount = types.Int64Value(int64(retryCount))
			}
			if redundancyInterval, ok := alertData["redundancy_interval"].(float64); ok {
				state.RedundancyInterval = types.Int64Value(int64(redundancyInterval))
			}
			if dataLimit, ok := alertData["data_limit"].(float64); ok {
				state.DataLimit = types.Int64Value(int64(dataLimit))
			}
			if sim, ok := alertData["sim"].(string); ok {
				var simInt int64
				fmt.Sscanf(sim, "%d", &simInt) //nolint:errcheck // parsing string to int
				state.SIM = types.Int64Value(simInt)
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AlertConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlertConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !plan.Action.IsNull() {
		updateReq["action"] = plan.Action.ValueInt64()
	}
	if !plan.Subject.IsNull() {
		updateReq["subject"] = plan.Subject.ValueString()
	}
	if !plan.Message.IsNull() {
		updateReq["message"] = plan.Message.ValueString()
	}
	if !plan.Email.IsNull() {
		updateReq["email"] = plan.Email.ValueString()
	}
	if !plan.SMTPConfigID.IsNull() {
		updateReq["smtp_config_id"] = plan.SMTPConfigID.ValueInt64()
	}
	if !plan.DeliveryRetry.IsNull() {
		updateReq["delivery_retry"] = plan.DeliveryRetry.ValueBool()
	}
	if !plan.RetryInterval.IsNull() {
		updateReq["retry_interval"] = plan.RetryInterval.ValueInt64()
	}
	if !plan.RetryCount.IsNull() {
		updateReq["retry_count"] = plan.RetryCount.ValueInt64()
	}
	if !plan.RedundancyInterval.IsNull() {
		updateReq["redundancy_interval"] = plan.RedundancyInterval.ValueInt64()
	}
	if !plan.DataLimit.IsNull() {
		updateReq["data_limit"] = plan.DataLimit.ValueInt64()
	}
	if !plan.SIM.IsNull() {
		updateReq["sim"] = plan.SIM.ValueInt64()
	}

	if len(updateReq) == 0 {
		resp.Diagnostics.AddError("No fields to update", "At least one field must be provided for update")
		return
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/alerts-configurations/%d", plan.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating alert configuration", fmt.Sprintf("Could not update alert configuration %d: %s", plan.ID.ValueInt64(), err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/alerts-configurations/%d", state.ID.ValueInt64()), nil); err != nil {
		resp.Diagnostics.AddError("Error deleting alert configuration", fmt.Sprintf("Could not delete alert configuration %d: %s", state.ID.ValueInt64(), err))
		return
	}
}

func (r *AlertConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AlertConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
