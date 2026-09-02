package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	// RMS v3: POST /devices/alerts-configurations
	// Returns {"success":true} — no id in response. Need list lookup by device_id + alert_type_id.
	createReq := map[string]interface{}{
		"device_id":     plan.DeviceID.ValueInt64(),
		"alert_type_id": plan.AlertTypeID.ValueInt64(),
		"action":        int(plan.Action.ValueInt64()),
		"message":       plan.Message.ValueString(),
	}

	if !plan.AlertSubtypeID.IsNull() {
		createReq["alert_subtype_id"] = plan.AlertSubtypeID.ValueInt64()
	}
	if !plan.Subject.IsNull() {
		createReq["subject"] = plan.Subject.ValueString()
	}
	if !plan.Email.IsNull() {
		createReq["email"] = plan.Email.ValueString()
	}
	if !plan.SMTPConfigID.IsNull() {
		createReq["smtp_config_id"] = plan.SMTPConfigID.ValueInt64()
	}
	if !plan.DeliveryRetry.IsNull() {
		createReq["delivery_retry"] = plan.DeliveryRetry.ValueBool()
	}
	if !plan.RetryInterval.IsNull() {
		createReq["retry_interval"] = plan.RetryInterval.ValueInt64()
	}
	if !plan.RetryCount.IsNull() {
		createReq["retry_count"] = plan.RetryCount.ValueInt64()
	}
	if !plan.RedundancyInterval.IsNull() {
		createReq["redundancy_interval"] = plan.RedundancyInterval.ValueInt64()
	}
	if !plan.DataLimit.IsNull() {
		createReq["data_limit"] = plan.DataLimit.ValueInt64()
	}
	if !plan.SIM.IsNull() {
		createReq["sim"] = plan.SIM.ValueInt64()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/devices/alerts-configurations", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating alert configuration", fmt.Sprintf("Could not create alert configuration: %s", err))
		return
	}

	// Post-create: list lookup by device_id + alert_type_id
	if plan.ID.IsUnknown() || (plan.ID.IsNull()) {
		if err := r.lookupAlertConfigByID(ctx, &plan); err != nil {
			resp.Diagnostics.AddError(
				"Error resolving alert configuration ID",
				fmt.Sprintf("Could not find created alert config for device %d: %s", plan.DeviceID.ValueInt64(), err),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// lookupAlertConfigByID fetches all alert configurations and finds the one matching device_id + alert_type_id.
// Uses exponential backoff retry (same pattern as lookupTagByID).
func (r *AlertConfigurationResource) lookupAlertConfigByID(ctx context.Context, plan *AlertConfigurationResourceModel) error {
	const maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(1<<uint(i)) * 200 * time.Millisecond
			time.Sleep(backoff)
		}

		// The client unwraps the RMS envelope, so the list decodes directly.
		var items []map[string]interface{}
		if err := r.client.Get(ctx, "/devices/alerts-configurations", map[string]string{
			"limit":     strconv.Itoa(listPageSize),
			"device_id": strconv.FormatInt(plan.DeviceID.ValueInt64(), 10),
		}, &items); err != nil {
			if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "forbidden") {
				return err
			}
			continue
		}

		targetDeviceID := plan.DeviceID.ValueInt64()
		targetAlertTypeID := plan.AlertTypeID.ValueInt64()
		for _, itemMap := range items {
			deviceID, ok := itemMap["device_id"].(float64)
			if !ok {
				continue
			}
			alertTypeID, ok := itemMap["alert_type_id"].(float64)
			if !ok {
				continue
			}
			if int64(deviceID) == targetDeviceID && int64(alertTypeID) == targetAlertTypeID {
				if id, ok := itemMap["id"].(float64); ok {
					plan.ID = types.Int64Value(int64(id))
				}
				// Hydrate all fields from list item (not just ID)
				plan.DeviceID = types.Int64Value(int64(deviceID))
				plan.AlertTypeID = types.Int64Value(int64(alertTypeID))
				if subID, ok := itemMap["alert_subtype_id"].(float64); ok {
					plan.AlertSubtypeID = types.Int64Value(int64(subID))
				}
				if action, ok := itemMap["action"].(float64); ok {
					plan.Action = types.Int64Value(int64(action))
				}
				if subject, ok := itemMap["subject"].(string); ok {
					plan.Subject = types.StringValue(subject)
				}
				if message, ok := itemMap["message"].(string); ok {
					plan.Message = types.StringValue(message)
				}
				if email, ok := itemMap["email"].(string); ok {
					plan.Email = types.StringValue(email)
				}
				if smtpID, ok := itemMap["smtp_config_id"].(float64); ok {
					plan.SMTPConfigID = types.Int64Value(int64(smtpID))
				}
				if retry, ok := itemMap["delivery_retry"].(bool); ok {
					plan.DeliveryRetry = types.BoolValue(retry)
				}
				if retryInterval, ok := itemMap["retry_interval"].(float64); ok {
					plan.RetryInterval = types.Int64Value(int64(retryInterval))
				}
				if retryCount, ok := itemMap["retry_count"].(float64); ok {
					plan.RetryCount = types.Int64Value(int64(retryCount))
				}
				if redundancyInterval, ok := itemMap["redundancy_interval"].(float64); ok {
					plan.RedundancyInterval = types.Int64Value(int64(redundancyInterval))
				}
				if dataLimit, ok := itemMap["data_limit"].(float64); ok {
					plan.DataLimit = types.Int64Value(int64(dataLimit))
				}
				if sim, ok := itemMap["sim"].(float64); ok {
					plan.SIM = types.Int64Value(int64(sim))
				}
				return nil
			}
		}
	}

	return fmt.Errorf("alert configuration not found after %d retries", maxRetries)
}

func (r *AlertConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GET /alerts-configurations/{id} returns the configuration itself; the
	// /devices/alerts-configurations/{id} path RMS never exposed.
	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/alerts-configurations/%d", state.ID.ValueInt64()), nil, &result); err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading alert configuration", fmt.Sprintf("Could not read alert configuration %d: %s", state.ID.ValueInt64(), err))
		return
	}

	if deviceID, ok := result["device_id"].(float64); ok {
		state.DeviceID = types.Int64Value(int64(deviceID))
	}
	if alertTypeID, ok := result["alert_type_id"].(float64); ok {
		state.AlertTypeID = types.Int64Value(int64(alertTypeID))
	}
	if alertSubtypeID, ok := result["alert_subtype_id"].(float64); ok {
		state.AlertSubtypeID = types.Int64Value(int64(alertSubtypeID))
	}
	action, diags := alertConfigInt64(result["action"], "action")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !action.IsNull() {
		state.Action = action
	}
	if subject, ok := result["subject"].(string); ok {
		state.Subject = types.StringValue(subject)
	}
	if message, ok := result["message"].(string); ok {
		state.Message = types.StringValue(message)
	}
	if email, ok := result["email"].(string); ok {
		state.Email = types.StringValue(email)
	}
	if smtpConfigID, ok := result["smtp_config_id"].(float64); ok {
		state.SMTPConfigID = types.Int64Value(int64(smtpConfigID))
	}
	if deliveryRetry, ok := result["delivery_retry"].(bool); ok {
		state.DeliveryRetry = types.BoolValue(deliveryRetry)
	}
	if retryInterval, ok := result["retry_interval"].(float64); ok {
		state.RetryInterval = types.Int64Value(int64(retryInterval))
	}
	if retryCount, ok := result["retry_count"].(float64); ok {
		state.RetryCount = types.Int64Value(int64(retryCount))
	}
	if redundancyInterval, ok := result["redundancy_interval"].(float64); ok {
		state.RedundancyInterval = types.Int64Value(int64(redundancyInterval))
	}
	if dataLimit, ok := result["data_limit"].(float64); ok {
		state.DataLimit = types.Int64Value(int64(dataLimit))
	}
	sim, diags := alertConfigInt64(result["sim"], "sim")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !sim.IsNull() {
		state.SIM = sim
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// alertConfigInt64 reads a field RMS returns either as a JSON number or as a
// decimal string. It yields a null value when the field is absent.
func alertConfigInt64(raw interface{}, field string) (types.Int64, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch v := raw.(type) {
	case float64:
		return types.Int64Value(int64(v)), diags
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			diags.AddError(
				fmt.Sprintf("Error parsing alert %s", field),
				fmt.Sprintf("Could not parse %s %q as a number: %s", field, v, err),
			)
			return types.Int64Null(), diags
		}
		return types.Int64Value(parsed), diags
	default:
		return types.Int64Null(), diags
	}
}

func (r *AlertConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlertConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// PUT /alerts-configurations/{id} takes the fields flat, with no wrapper.
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

	// The delete operation is scoped to the owning device.
	if state.DeviceID.IsNull() || state.DeviceID.IsUnknown() {
		resp.Diagnostics.AddError(
			"Cannot delete alert configuration",
			fmt.Sprintf("Alert configuration %d has no device_id in state, and RMS deletes an alert "+
				"configuration only through DELETE /devices/{device_id}/alerts-configurations/{alert_id}. "+
				"Re-import the resource so device_id is recorded.", state.ID.ValueInt64()),
		)
		return
	}

	deletePath := fmt.Sprintf("/devices/%d/alerts-configurations/%d", state.DeviceID.ValueInt64(), state.ID.ValueInt64())
	if err := r.client.Delete(ctx, deletePath, nil); err != nil {
		if errors.Is(err, api.ErrNotFound) {
			return
		}
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
