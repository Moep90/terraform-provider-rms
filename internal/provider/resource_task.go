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
	var data TaskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":       data.Name.ValueString(),
		"task_type":  data.TaskType.ValueString(),
		"company_id": data.CompanyID.ValueInt64(),
	}

	if !data.Description.IsNull() {
		createReq["description"] = data.Description.ValueString()
	}
	if !data.Payload.IsNull() {
		createReq["payload"] = data.Payload.ValueString()
	}
	if !data.ScheduledAt.IsNull() {
		createReq["scheduled_at"] = data.ScheduledAt.ValueString()
	}
	if !data.TaskGroupID.IsNull() {
		createReq["task_group_id"] = data.TaskGroupID.ValueInt64()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/tasks", createReq, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error creating task",
			fmt.Sprintf("Could not create task: %s", err),
		)
		return
	}

	id, ok := result["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing task ID",
			"Could not parse task ID from API response",
		)
		return
	}
	data.ID = types.Int64Value(int64(id))

	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	} else {
		data.Description = types.StringValue("")
	}
	if taskType, ok := result["task_type"].(string); ok {
		data.TaskType = types.StringValue(taskType)
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

func (r *TaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/tasks/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		if err.Error() == "API error 404: 404 Not Found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading task",
			fmt.Sprintf("Could not read task %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	if id, ok := result["id"].(float64); ok {
		data.ID = types.Int64Value(int64(id))
	}
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
	if taskType, ok := result["task_type"].(string); ok {
		data.TaskType = types.StringValue(taskType)
	}
	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}
	if taskGroupID, ok := result["task_group_id"].(float64); ok {
		data.TaskGroupID = types.Int64Value(int64(taskGroupID))
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
	if err := r.client.Put(ctx, fmt.Sprintf("/tasks/%d", data.ID.ValueInt64()), updateReq, &result); err != nil {
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
	if err := r.client.Delete(ctx, fmt.Sprintf("/tasks/%d", data.ID.ValueInt64()), &result); err != nil {
		if err.Error() == "API error 404: 404 Not Found" {
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
