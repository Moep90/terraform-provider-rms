package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &TaskResource{}
var _ resource.ResourceWithImportState = &TaskResource{}

func NewTaskResource() resource.Resource {
	return &TaskResource{}
}

type TaskResource struct {
	client *api.Client
}

type TaskResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	TaskType    types.String `tfsdk:"task_type"`
	Status      types.String `tfsdk:"status"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	TaskGroupID types.Int64  `tfsdk:"task_group_id"`
	Payload     types.String `tfsdk:"payload"`
	ScheduledAt types.String `tfsdk:"scheduled_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *TaskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_task"
}

func (r *TaskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Task. Tasks are used to send commands or configurations to devices.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the task.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the task.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the task.",
			},
			"task_type": schema.StringAttribute{
				Required:    true,
				Description: "The type of task (e.g., 'reboot', 'config_update', 'firmware_upgrade').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the task (e.g., 'pending', 'running', 'completed', 'failed').",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID that owns this task.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"task_group_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The task group ID this task belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"payload": schema.StringAttribute{
				Optional:    true,
				Description: "JSON payload containing task-specific parameters.",
			},
			"scheduled_at": schema.StringAttribute{
				Optional:    true,
				Description: "Scheduled execution time in ISO8601 format.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp in ISO8601 format.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp in ISO8601 format.",
			},
		},
	}
}

func (r *TaskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *TaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// RMS defines GET on /devices/tasks and nothing else: there is no create
	// operation for an individual task. Fail here rather than issuing a request
	// that cannot succeed, or worse, recording state for an object RMS never
	// made.
	resp.Diagnostics.AddError(
		"rms_task cannot be created",
		"The RMS API exposes no create operation for individual tasks "+
			"(POST /devices/tasks is not defined). Create a task group with "+
			"rms_task_group, which RMS does support through POST /devices/tasks/groups, "+
			"or use `terraform import` to manage a task that already exists.",
	)
}

func (r *TaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// RMS exposes no GET /devices/tasks/{id}; the task has to be found in the
	// collection.
	result, err := findInList(ctx, r.client, "/devices/tasks", data.ID.ValueInt64())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading task",
			fmt.Sprintf("Could not read task %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
	if taskType, ok := result["type"].(string); ok {
		data.TaskType = types.StringValue(taskType)
	}
	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}
	if groupID, ok := result["group_id"].(float64); ok {
		data.TaskGroupID = types.Int64Value(int64(groupID))
	}
	if payload, ok := result["payload"].(string); ok {
		data.Payload = types.StringValue(payload)
	}
	if scheduledAt, ok := result["scheduled_at"].(string); ok {
		data.ScheduledAt = types.StringValue(scheduledAt)
	}
	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}
	if updatedAt, ok := result["updated_at"].(string); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Description.IsNull() {
		updateReq["description"] = data.Description.ValueString()
	}
	if !data.Payload.IsNull() {
		updateReq["payload"] = data.Payload.ValueString()
	}
	if !data.ScheduledAt.IsNull() {
		updateReq["scheduled_at"] = data.ScheduledAt.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/devices/tasks/%d", data.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error updating task",
			fmt.Sprintf("Could not update task %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	} else {
		data.Status = types.StringValue("pending")
	}
	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	} else {
		data.CreatedAt = types.StringValue("2024-01-01T00:00:00Z")
	}
	if updatedAt, ok := result["updated_at"].(string); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	} else {
		data.UpdatedAt = types.StringValue("2024-01-01T00:00:00Z")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Delete(ctx, fmt.Sprintf("/devices/tasks/%d", data.ID.ValueInt64()), &result); err != nil {
		if errors.Is(err, api.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting task",
			fmt.Sprintf("Could not delete task %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}
}

func (r *TaskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// "id" is an Int64 attribute, so the raw string import ID cannot be passed
	// through unconverted.
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected a numeric task ID, got %q.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
